'use strict';

const WASM_URL = 'wasm.wasm';

var wasm;

function clearCanvas() {
  wasm.exports.clearCanvas();
}

function drawLine() {
  wasm.exports.drawLine();
}

function init() {
  const go = new Go();
  if ('instantiateStreaming' in WebAssembly) {
    WebAssembly.instantiateStreaming(fetch(WASM_URL), go.importObject).then(function (obj) {
      wasm = obj.instance;
      go.run(wasm);
      clearCanvas();
    })
  } else {
    fetch(WASM_URL).then(resp =>
      resp.arrayBuffer()
    ).then(bytes =>
      WebAssembly.instantiate(bytes, go.importObject).then(function (obj) {
        wasm = obj.instance;
        go.run(wasm);
        clearCanvas();
      })
    )
  }



}

init();
