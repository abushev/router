# router

Routing incoming requests, and saving general statistics, which is available on dstatus action call.

* Example:
  ```
  router := router.Router{
    Level:  1,
    Offset: 0,
    Dstatus: router.NewDstatus(router.DstatusData{
      Name: "Router example",
      Port: "8080",
    }),
    Routes: map[string]router.ActionStruct{
      "action1": router.ActionStruct{Fn: action1Func},
      "action2": router.ActionStruct{Fn: action2Func},
    },
  }
  router.Dstatus.Additional = func() string {
    return "Some additional data to be shown in dstatus"
  }
  http.HandleFunc("/", router.Handler)
   ```
