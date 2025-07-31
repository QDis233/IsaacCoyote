package isaac

import (
	"encoding/json"
	"os"
	"sync"
)

type ResourceManager struct {
	items     map[int]ItemDetail
	itemIndex map[string]int

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
	itemData, err := os.ReadFile("resources/items.json")
	if err != nil {
		return err
	}
	ItemDetails := make(map[int]ItemDetail)
	err = json.Unmarshal(itemData, &ItemDetails)
	if err != nil {
		return err
	}

	for id, item := range ItemDetails {
		m.itemIndex[item.Name] = id
	}
	m.items = ItemDetails

	return nil
}

func (m *ResourceManager) GetItemByName(itemName string) (ItemDetail, error) {
	if itemID, ok := m.itemIndex[itemName]; ok {
		if item, ok := m.items[itemID]; ok {
			return item, nil
		}
	}
	return ItemDetail{}, NoSuchItemError{Message: "No such item: " + itemName}
}

func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		items:     make(map[int]ItemDetail),
		itemIndex: make(map[string]int),
	}
}
