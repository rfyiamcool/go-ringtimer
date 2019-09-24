# go-ringtimer

![](timewheel.png)

## Usage:

```
tw, err := NewTimeWheel(time.Second, 60)
if err != nil {
    t.Error(err)
}
tw.Start()

entry, _ := tw.AfterFunc(1 * time.Second, func() {
})

<- entry.C
entry.Reset(2 * time.Second)
entr.Stop()

tw.Sleep(1 * time.Second)
<- tw.After(1 * time.Second)
```

## Performance:

insert 120w timer event for 60 second, call time cost 61s.

**see example/main.go**