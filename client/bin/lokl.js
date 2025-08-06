#! /usr/bin/env node

const { spawn } = require('child_process')
const path = require('path')

const platform = process.platform;
let binName;

if(platform == 'win32'){
    binName = 'lokl-cli.exe'
}
else if(platform == 'linux'){
    binName = 'lokl-cli'
}
else{
    console.log('Not supported platform')
    process.exit(1);
}

const binPath = path.join(__dirname, binName);
const child = spawn(binPath, [], {stdio:'inherit'})
child.on('exit', (code) => {
    process.exit(code);
})