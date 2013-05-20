Copyright 2013 Vlad Didenko. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the header of the fanout.go file.

Package pumps is a collection of channel manipulators. Contained objects expected to communicate via channels, including meta-communication.

Here is an example using the FanOut object for handling errors:

In the main package
-------------------

    import pumps

    // in the init()
    var errCast = pumps.MakeFanOut(1)

    // in the main()
    defer func(){ errCast.Post <- nil }()

In a syslog forwarder setup
---------------------------

    fwdToSysLog := make(chan error, 100)
    errCast.Outs <- fwdToSysLog
    go func() {
        for err := range fwdToSysLog { ...... }
    }()

In an operational console handler setup
---------------------------------------

    peerConnErrors := make(chan *PeerConnError, 100)
    errCast.Outs <- peerConnErrors

    hardwareErrors := make(chan *HardwareError, 100)
    errCast.Outs <- hardwareErrors

Generating and posting an error
-------------------------------

    hwError := &ops.HardwareError{
        devicePath,
        errorCode,
        errors.New("CRC failed"),
    }
    errCast.Post <- hwError

The error will be fowarded to `fwdToSysLog` and `hardwareErrors` channels, but not `peerConnErrors` (assuming `PeerConnError` is not assignable to `HardwareError`).
