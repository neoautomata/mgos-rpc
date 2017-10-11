# mgos-rpc

A simple command line interface to send [Mongoose-OS](http://www.mongoose-os.com) RPCs over either MQTT or Web Socket.

## Install

```shell
$ go get -u github.com/neoautomata/mgos-rpc
```

## Usage

```shell
$ mqttrpc [--address] tcp://user:password@mqtt-borker:1883#mgos-device-id [--method] RPC.Method [--print_resp true|false] [arg1=val1] [arg2=val2] [arg3=val3] ...

$ wsrpc [--address] host_or_ip [--method] RPC.Method [--print_resp true|false] [arg1=val1] [arg2=val2] [arg3=val3] ...
```
