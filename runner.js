#! /usr/bin/env node
const { runClib, sendClib } = require('.');

let nodePort = 0;
function fromClib(port, str, replier) {
  console.error(`runner inbound ${port} ${str}`);
  setTimeout(() => {
    console.log(sendClib(clibPort, `runner called back with ${str}`));
    replier.resolve(`runner replied ${str}`);
  }, 2000);
}

nodePort += 1;
const clibPort = runClib(nodePort, fromClib, ['Hello', 'Agoric', 'world!']);
