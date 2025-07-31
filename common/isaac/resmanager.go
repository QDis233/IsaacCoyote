package isaac

import (
	"encoding/json"
	"os"
	"sync"
)

type ResourceManager struct {
	collectibles      map[int]ItemDetail
	collectiblesIndex map[string]int

	resLock sync.Mutex
}

func (m *ResourceManager) LoadResources() error {
	m.resLock.Lock()
	defer m.resLock.Unlock()

	err := m.praseItemDetails()
	if err != nil {
		return err
	}
	return nil
}

func (m *ResourceManager) praseItemDetails() error {
	itemData, err := os.ReadFile("resources/collectibles.json")
	if err != nil {
		return err
	}
	ItemDetails := make(map[int]ItemDetail)
	err = json.Unmarshal(itemData, &ItemDetails)
	if err != nil {
		return err
	}

	for id, item := range ItemDetails {
		m.collectiblesIndex[item.Name] = id
	}
	m.collectibles = ItemDetails

	return nil
}

func (m *ResourceManager) GetItemByName(itemName string) (ItemDetail, error) {
	if itemID, ok := m.collectiblesIndex[itemName]; ok {
		if item, ok := m.collectibles[itemID]; ok {
			return item, nil
		}
	}
	return ItemDetail{}, NoSuchItemError{Message: "No such item: " + itemName}
}

func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		collectibles:      make(map[int]ItemDetail),
		collectiblesIndex: make(map[string]int),
	}
}
