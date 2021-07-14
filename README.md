# redisobj
Redis Model Mapping

[![Open in Visual Studio Code](https://open.vscode.dev/badges/open-in-vscode.svg)](https://open.vscode.dev/wspowell/redisobj)

This is an experiment to see how redis can be mapped to models in order to abstract away the majority of redis calls.
Storage is backed by go-redis.

# Usage

## Store
The redis store must be created with a go-redis UniversalClient. The Store will lazy initialize struct data as they are used. 
The store is thread safe.
```
// This value will be used in the following examples.
objStore := redisobj.NewStore(redisClient)
```

## Saving golang structs
All keys stored with redisobj are co-located on redis nodes by utilizing redis hash tags. 

### Singleton
A singleton struct can be saved to redis with no modifications. The struct can then be read simply by read the data into the same struct type.
```
type Singleton struct {
  GlobalValue string
}

singleton := Singleton{
  GlobalValue: "value",
}

err := objStore.Write(singleton)
```
This will create redis keys for Singleton as:
* {redisobj:Singleton} - hash

Writing the same object with different values will override the data stored in redis with the new values.

Since the key is not based on any data, that means it can retrieved using any instance of Singleton.
```
err := objStore.Write(singleton)
```

### Keyed Data
Objects that are based on keys or IDs can be used by providing a struct field with the struct tag "key".
```
type Item struct {
  Id       string            `redisobj:"key"`
  Value    int
  Metadata map[string]string 
}

item123 := Item{
  Id:    "123",
  Value: "item123",
}

item999 := Item{
  Id:       "999",
  Value:    "item999",
  Metadata: map[string]string{
    "CreatedBy": "admin",
  },
}

err := objStore.Write(item123)
err := objStore.Write(item999)
```
This will create two Items in redis under two different keys:
* {redisobj:Item:123} - hash
* {redisobj:Item:999} - hash

However, since Item 999 has a map, it will be stored under another key to avoid collisions:
* {redisobj:Item:999}.Metadata - hash

The keyed items can be read from redis, but the key value must be supplied.
```
item123 := Item{
  Id: "123",
}

item999 := Item{
  Id: "999",
}

// Reads the two items from redis into the structs.
err := objStore.Read(&item123)
err := objStore.Read(&item999)

// If an item does not exist an error will be returned.
if err != nil && errors.Is(err, redisobj.ErrObjectNotFound) {
  // Handle not found error.
}
```

## Nested Data and Keys
Nested structs may be stored in one of a few configurations.
1. Neither struct has a key
  * The keys use the same prefix and will be found in the same hash slot
```
// {redisobj:Singleton}
type Singleton struct {
  Value string
  Data Metadata
}

// {redisobj:Singleton}:Metadata
type Metadata struct {
  Info string
}
```
2. The root struct has a key
  * The keys use the same prefix and will be found in the same hash slot by key value.
```
// {redisobj:Item:<Id>}
type Item struct {
  Id    string   `redisobj:"key"`
  Value string 
  Data  Metadata
}

// {redisobj:Item:<Id>}:Metadata
type Metadata struct {
  Info string
}
```
3. The nested struct has a key
  * The struct have different key prefixes and may be found in different hash slots
```
// {redisobj:Item:<Id>}
type Item struct {
  Id        string `redisobj:"key"`
  Value     string 
  ItemGroup Group
}

// {redisobj:Group:<Id>}
type Group struct {
  Id    string `redisobj:"key"`
  Value string
}
```
It is important to note that nested keys can be accessed from redis independent from one another.
```
item := Item{
  Id:        "123",
  Value:     "TV",
  ItemGroup: Group{
    Id:    "999",
    Value: "Electronics",
  },
}

...

// And then later accessed independently.

item := Item{
  Id: "123",
}

group := Group{
  Id: "999",
}

// Reads the two items from redis into the structs.
err := objStore.Read(&item)  // Item will include the Group data
err := objStore.Read(&group)
```

# Benchmarks
TODO Show how redisobj compares to redis commands.
