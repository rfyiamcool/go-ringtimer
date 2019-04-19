# go-timewheel

![](timewheel.png)

## Usage:

```
tw, err := NewTimeWheel(time.Second, 60)
if err != nil {
    t.Error(err)
}
tw.Start()

tw.AfterFunc(1 * time.Second, func() {
})

tw.Sleep(1 * time.Second)
<- tw.After(1 * time.Second)
```

## TO DO List:

* like go' newTimer
* like go' newTick

## Performance:

insert 120w timer event for 60 second, call time cost 61s.

**see example/main.go**