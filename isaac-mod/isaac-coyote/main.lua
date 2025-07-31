local mod                = RegisterMod("IsaacCoyote", 1)

-- Includes
local json               = require("json")

---Constants
local HEARTBEAT_INTERVAL = 5 * 60 -- 5 seconds
local UPDATE_FREQUENCY   = 15     -- 15 frames: 1/4 seconds
local COYOTE_CALLBACKS   = {
    C_ON_INDICATOR_UPDATE = "update_indicator",
    C_ON_CONNECT = "connect",
    C_ON_HEARTBEAT = "heartbeat",
}

---variables
local modSettings        = {
    IndicatorOffsetX = 4,
    IndicatorOffsetY = 190,
    IndicatorSize = 10
}

local localPlayerRNG
local collectiblesList   = {}

local isPrevGameExited   = false
local isPrevGameLiving   = true

local isConnected        = false
local isRecviedHeartbeat = false
local heartbeatTimer     = HEARTBEAT_INTERVAL
local dataTable
local indicatorData
local font               = Font()
local game               = Game()
font:Load("font/cjk/lanapixel.fnt")

--- func
local function initConfigMenu()
    if ModConfigMenu == nil then
        return
    end


    ModConfigMenu.AddSetting(
        "IsaacCoyote",
        "Indicator",
        {
            Type = ModConfigMenu.OptionType.NUMBER,
            CurrentSetting = function()
                return modSettings.IndicatorOffsetX
            end,
            Display = function()
                return "Indicator Offset X: " .. modSettings.IndicatorOffsetX
            end,
            OnChange = function(n)
                modSettings.IndicatorOffsetX = n
            end,
            Info = { "Indicator Offset X" }
        }
    )

    ModConfigMenu.AddSetting(
        "IsaacCoyote",
        "Indicator",
        {
            Type = ModConfigMenu.OptionType.NUMBER,
            CurrentSetting = function()
                return modSettings.IndicatorOffsetY
            end,
            Display = function()
                return "Indicator Offset Y: " .. modSettings.IndicatorOffsetY
            end,
            OnChange = function(n)
                modSettings.IndicatorOffsetY = n
            end,
            Info = { "Indicator Offset Y" }
        }
    )

    ModConfigMenu.AddSetting(
        "IsaacCoyote",
        "Indicator",
        {
            Type = ModConfigMenu.OptionType.NUMBER,
            CurrentSetting = function()
                return modSettings.IndicatorSize
            end,
            Display = function()
                return "Indicator Size: " .. modSettings.IndicatorSize/10
            end,
            OnChange = function(n)
                modSettings.IndicatorSize = n
            end,
            Info = { "Indicator Size" }
        }
    )
end

local function newDataTable()
    local object = {
        rawData = "",
        data = {
            send = {},
            receive = {},
        },
        newMessages = {

        },
        callbacks = {
            [COYOTE_CALLBACKS.C_ON_HEARTBEAT] = {

            },
            [COYOTE_CALLBACKS.C_ON_INDICATOR_UPDATE] = {

            },
            [COYOTE_CALLBACKS.C_ON_CONNECT] = {

            }
        }
    }

    function object.PushMessage(eventObj)
        eventObj.frameCount = Isaac.GetFrameCount()
        object.newMessages[#object.newMessages + 1] = eventObj
    end

    function object.WriteTable()
        object.data.send = object.newMessages
        mod:SaveData(json.encode(object.data))
    end

    function object.updateData()
        if not pcall(object._loadData) then
            return
        end

        if not object.data or not object.data.receive then
            return
        end

        if #object.data.receive > 0 then
            for _, event in ipairs(object.data.receive) do
                if object.callbacks[event.type] then
                    for _, callback in ipairs(object.callbacks[event.type]) do
                        callback(event.message)
                    end
                end
            end
        end
        object.data.receive = {}
    end

    function object.RegisterCallback(eventType, callback)
        if not object.callbacks[eventType] then
            return
        end
        table.insert(object.callbacks[eventType], callback)
    end

    function object._loadData()
        object.rawData = Isaac.LoadModData(mod)
        object.data = json.decode(object.rawData)
    end

    return object
end

local function newEventMsg(eventType, eventData)
    return {
        type = "event",
        message = {
            type = eventType,
            data = eventData,
        }
    }
end

local function newHeartbeatMsg()
    return {
        type = "heartbeat",
        message = {},
    }
end

local function getPlayerRNG(player)
    if player and player:ToPlayer() then
        return player:GetCollectibleRNG(CollectibleType.COLLECTIBLE_SAD_ONION):GetSeed()
    end
end

local function getLocalPlayer()
    if localPlayerRNG then
        local playerNums = game.GetNumPlayers(game)
        for i = 0, playerNums - 1 do
            local player = game:GetPlayer(i)
            if player and getPlayerRNG(player) == localPlayerRNG then
                return player
            end
        end
    end
    return game:GetPlayer(0) --- HACK: Fallback ?
end

local function checkConnection()
    if isConnected then
        heartbeatTimer = heartbeatTimer - 1
        if heartbeatTimer <= 0 then
            if not isRecviedHeartbeat then
                isConnected = false
                indicatorData = {
                    strengthA = 0,
                    strengthB = 0,
                }
                return
            end
            heartbeatTimer = HEARTBEAT_INTERVAL
            isRecviedHeartbeat = false
        end
    end
end

local function RenderIndicator()
    local size = modSettings.IndicatorSize / 10
    if not isConnected then
        font:DrawStringScaledUTF8(
            "等待连接...",
            modSettings.IndicatorOffsetX,
            modSettings.IndicatorOffsetY,
            size,
            size,
            KColor(1, 0, 0, 0.8),
            0,
            false
        )
        return
    end

    font:DrawStringScaledUTF8(
        "当前电量:",
        modSettings.IndicatorOffsetX,
        modSettings.IndicatorOffsetY,
        size,
        size,
        KColor(1, 1, 1, 0.8),
        0,
        false
    )
    font:DrawStringScaledUTF8(
        string.format("A: %d B:%d", indicatorData.strengthA, indicatorData.strengthB),
        modSettings.IndicatorOffsetX + (size) * 2,
        modSettings.IndicatorOffsetY + (size) *12,
        size,
        size,
        KColor(1, 1, 1, 0.8),
        0,
        false
    )
end

--- Conn Callbacks
local function onConnect(data)
    heartbeatTimer = HEARTBEAT_INTERVAL
    isConnected = true
end

local function onHeartbeat(data)
    isRecviedHeartbeat = true
    dataTable.PushMessage(newHeartbeatMsg())
end

local function onUpdateIndicatorData(data)
    indicatorData.strengthA = data.strengthA or 0
    indicatorData.strengthB = data.strengthB or 0
end


--- Mod Callbacks
function mod:onRender()
    local frameCount = Isaac.GetFrameCount()
    if frameCount % UPDATE_FREQUENCY == 0 then
        dataTable.updateData()
    end
    checkConnection()

    if isConnected then
        if frameCount % UPDATE_FREQUENCY == 0 then
            player = getLocalPlayer()
            if player and player:GetHearts() > 0 then
                dataTable.PushMessage(newEventMsg("PlayerInfoUpdateEvent", {
                    health = player:GetHearts(),
                    maxHealth = player:GetMaxHearts(),
                }))
            end
            dataTable.WriteTable()
            dataTable.newMessages = {}
        end
    end
    RenderIndicator()
end

function mod:regLocalPlayerRNG(player)
    if player and player.FrameCount == 1 then
        localPlayerRNG = getPlayerRNG(player)
    end
end

function mod:checkNewCollectible()
    player = getLocalPlayer()
    if not player then
        return
    end

    -- local itemData = player.QueuedItem.Item
    -- if itemData ~= nil and itemData.IsCollectible(itemData) then
    --     if not queuedItemList[itemData.ID] then
    --         queuedItemList[itemData.ID] = itemData
    --     end
    -- else
    --     for id, _ in pairs(queuedItemList) do
    --         local itemData = queuedItemList[id]
    --         local eventData = {
    --             name = itemData.Name,
    --             id = itemData.ID,
    --             quality = itemData.Quality,
    --         }
    --         dataTable.PushMessage(newEventMsg("NewCollectibleEvent", eventData))
    --         queuedItemList[id] = nil
    --     end
    -- end

    local itemData = player.QueuedItem.Item
    if itemData ~= nil and itemData.IsCollectible(itemData) then
        if not collectiblesList[itemData.ID] then
            local eventData = {
                name = itemData.Name,
                id = itemData.ID,
                quality = itemData.Quality,
            }
            dataTable.PushMessage(newEventMsg("NewCollectibleEvent", eventData))
            collectiblesList[itemData.ID] = itemData
        end
    end
end

function mod:onPlayerDamage(entity, damage, flags, source, countdown)
    if not isConnected then
        return
    end

    local player = entity:ToPlayer()
    if player and getPlayerRNG(player) == localPlayerRNG then
        local eventData = {
            playerName = player:GetName(),
            damage = damage,
            flags = flags,
            source = source.Type,
        }
        dataTable.PushMessage(newEventMsg("PlayerHurtEvent", eventData))
    end
end

function mod:onPlayerDeath(entity)
    if not isConnected then
        return
    end


    local player = entity:ToPlayer()
    if player and getPlayerRNG(player) == localPlayerRNG then
        dataTable.PushMessage(newEventMsg("PlayerDeathEvent", {}))
    end
end

function mod:onExit()
    isPrevGameExited = true
    dataTable.PushMessage(newEventMsg("GameExitEvent", {}))
end

function mod:onGameEnd()
    isPrevGameLiving = false
    dataTable.PushMessage(newEventMsg("GameEndEvent", {}))
end

function mod:onGameStarted(isContinue)
    if not isContinue then
        if isPrevGameExited and isPrevGameLiving then
            dataTable.PushMessage(newEventMsg("ManualRestartEvent", {}))
        end
    end
    collectiblesList = {}
    dataTable.PushMessage(newEventMsg("GameStartEvent", { isContinue = isContinue }))
    isPrevGameExited = false
    isPrevGameLiving = true
end

---Main
---clear the save data on mod initialization
mod:SaveData('{"send":[],"receive":[]}')
initConfigMenu()
dataTable = newDataTable()
indicatorData = {
    strengthA = 0,
    strengthB = 0,
}

dataTable.RegisterCallback(COYOTE_CALLBACKS.C_ON_CONNECT, onConnect)
dataTable.RegisterCallback(COYOTE_CALLBACKS.C_ON_HEARTBEAT, onHeartbeat)
dataTable.RegisterCallback(COYOTE_CALLBACKS.C_ON_INDICATOR_UPDATE, onUpdateIndicatorData)

mod:AddCallback(ModCallbacks.MC_POST_RENDER, mod.onRender)

mod:AddCallback(ModCallbacks.MC_POST_RENDER, mod.checkNewCollectible)
mod:AddCallback(ModCallbacks.MC_POST_PEFFECT_UPDATE, mod.regLocalPlayerRNG)

mod:AddCallback(ModCallbacks.MC_ENTITY_TAKE_DMG, mod.onPlayerDamage)
mod:AddCallback(ModCallbacks.MC_POST_ENTITY_KILL, mod.onPlayerDeath)

mod:AddCallback(ModCallbacks.MC_PRE_GAME_EXIT, mod.onExit)
mod:AddCallback(ModCallbacks.MC_POST_GAME_END, mod.onGameEnd)
mod:AddCallback(ModCallbacks.MC_POST_GAME_STARTED, mod.onGameStarted)
