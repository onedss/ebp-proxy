#!/bin/bash
CWD=$(cd "$(dirname $0)";pwd)
"$CWD"/ebp-proxy install
"$CWD"/ebp-proxy start
