package main

import (
    "encoding/json"
    "os"
)

type omit bool

type Value interface{}

type CacheItem struct {
    Key    string `json:"key"`
    MaxAge int    `json:"cacheAge"`
    Value  Value  `json:"cacheValue"`
}

func NewCacheItem() (*CacheItem, error) {
    i := &CacheItem{}
    return i, json.Unmarshal([]byte(`{
      "key": "foo",
      "cacheAge": 1234,
      "cacheValue": {
        "nested": true
      }
    }`), i)
}

func main() {
    item, _ := NewCacheItem()

    json.NewEncoder(os.Stdout).Encode(struct {
        *CacheItem

        // Omit bad keys
        OmitMaxAge omit `json:"cacheAge,omitempty"`
        OmitValue  omit `json:"cacheValue,omitempty"`

        // Add nice keys
        MaxAge int    `json:"max_age"`
        Value  *Value `json:"value"`
    }{
        CacheItem: item,

        // Set the int by value:
        MaxAge: item.MaxAge,

        // Set the nested struct by reference, avoid making a copy:
        Value: &item.Value,
    })
}
