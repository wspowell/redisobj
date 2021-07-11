# redisobj
Redis Model Mapping

[![Open in Visual Studio Code](https://open.vscode.dev/badges/open-in-vscode.svg)](https://open.vscode.dev/wspowell/redisobj)

This is an experiment to see how redis can be mapped to models in order to abstract away the majority of redis calls.
Storage is backed by go-redis.

# Usage
Structs can be saved to redis by tagging fields as redis fields.
Supported struct tags:
* redisValue - a redis value backed by GET/SET commands
  * If the struct tag value ends with ",key" then this field will be used as a key value when generating the redis key.

# Examples

## Single values with key
```
// Item that stores two values in redis:
// * redisobj:item:<id>:id
// * redisobj:item:<id>:value
type Item struct {
  Id    string `redisValue:"id,key"`
  Value int    `redisValue:"value"`
}

objStore := redisobj.NewStore(redisClient)

// Write an Item with a given key.
anItem := Item{
  Id: "abc123",
  Value: 999,
}
err := objStore.Write(anItem)

// ...

// Read the item back using the same key.
// If Id is not set, then the correct Item will not be found.
retrievedItem := Item{
  Id: "abc123",
}
err := objStore.Read(&retrievedItem)

```
