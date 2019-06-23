This is just simple play around code, to learn what TinyGo
can do with the html5 canvas at present. :smile:

The running version of this is here:

&nbsp; &nbsp; https://justinclift.github.io/tinygo_canvas2/

To compile the WebAssembly file:

    $ tinygo build -target wasm -gc conservative -no-debug -o docs/wasm.wasm wasm.go
