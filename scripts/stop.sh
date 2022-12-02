#!/bin/bash
CWD=$(cd "$(dirname $0)";pwd)
"$CWD"/ebp-proxy stop
"$CWD"/ebp-proxy uninstall
