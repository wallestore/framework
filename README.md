framework
=============

## example

### set/get config
```javascript
//set config
func SetConfig(conf interface{})

//get config
func GetConfig() interface{}
```

### add heartbeat server
```javascript
framework.SetAppName("heartbeat")
framework.Heartbeat(10*time.Second, Func())
framework:start()
```