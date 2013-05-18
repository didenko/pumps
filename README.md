Package pumps is a collection of channel manipulators. Contained objects expected to communicate via channels, including meta-communication.

Here is an example using the FanOut object for handling errors:

In the main package
-------------------

    import pumps

    // in the init()
    var errCast = pumps.MakeFanOut(1)

    // in the main()
    defer func(){ errCast.Post <- nil }()

In a syslog forwarder
---------------------

    fwdToSysLog := make(chan error, 100)
    errCast.Outs <- fwdToSysLog
    go func() {
        for err := range fwdToSysLog { ...... }
    }()

In an operational console handler: