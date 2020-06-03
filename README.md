# Node.js Clib Bridge

This package is a template for writing a Node.js addon that communicates via function calls into a multithreaded C library.

The example, compiled with `make` and run with `./runner.js` shows how a Golang shared library can be used from Node.js without running into synchronization problems.  Downcalls to Golang are synchronous (they return when the Golang thread returns), and upcalls are asynchronous (they are enqueued and dispatched to the Node.js event queue, returning to Golang only when the returner's `.resolve` or `.reject` callbacks are called).

The C API exposed by clib.go is an example of what you will probably need to do in your own projects.  Note that "ports" are increasing integers for naming the endpoints in the runner and the clib, and to which pending upcall the resolve/reject is directed.

Adopting these conventions was driven by the mismatch between Golang (synchronous, multithreaded) and Node.js (asynchronous, event-driven), and it is probably useful for other languages that can compile into a C shared library.  Not too useful outside of Node.js, though.  Maybe later.

Have fun,
Michael FIG <mfig@agoric.com>, 2020-06-02
