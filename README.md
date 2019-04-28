This is just simple play around code, to learn what TinyGo
can do with the html5 canvas at present. :smile:

The running version of this is here:

&nbsp; &nbsp; https://justinclift.github.io/tinygo_canvas_test1/

So far, most stuff has "just worked", but the current lack of
a Garbage Collector (GC) for TinyGo's WebAssembly output has
turned out to be a blocker (for now).

To compile the WebAssembly file:

    $ tinygo build -o wasm.wasm -no-debug -target wasm wasm.go
