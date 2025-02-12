# Ringin: Dial-an-executable

Ringin listens on a modem and executes a program on incoming calls.

# Usage

First, make sure you have a working modem.

Edit ringin.toml to configure the modem

Example:

```toml
[serial]
  port = "/dev/ttyUSB0"
  baud_rate = 9600
  data_bits = 8
  parity = "N"
  stop_bits = 1
[modem]
   init_commands = ["ATH0", "AT&F0","ATV0", "ATX0", "ATE","ATS0=0","ATS6=10", "AT&D0"]
[program]
  command = "fortune"
  args = []
```

Start:

```
ringin -config ringin.toml
```

# License
Ringin is licensed under the MIT license.

Copyright 2025 Morgan Gangwere ("indrora") morgan.gangwere@gmail.com

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.