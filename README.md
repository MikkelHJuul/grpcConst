# gRPC constant

`grpcConst` is a slim reflection based tool to populate the client side data of gRPC calls without having to send that data in every call. This is useful to reduce the data transported between services where the stream messages share data. 

An example could be for spatiotemporal data where you may send several thousand data points that all share the same location, because a user queried a specific datatype (e.g temperature), at a specific location. And want to have a timeseries displayed. This data may share location, maybe a location name and you could have other fields that are added as convenience to the data, but those fields are constant across the request. 

## Specification
For a given `message` returned by an RPC server side stream. The server sends a defaulting proto marshal'ed base64 URLencoded header `x-grpc-const` this object of the same type of the `message` will be the default values for the client side to add to all messages. 

The client side does the most of the work, decoding, unmarshaling and padding the `message` with this data. 

## Overriding
Any `message` sent with a value in the same place as the default constant `message` 
will override the default.  
You cannot override the default by setting the value to 0, null, empty string or empty list as they are considered empty, you cannot set these values as default,  and that wouldn't make sense as they are default already. This is an important limitation; 
**if you expect to be able to send actual data of value 0, don't set a default on that value! This is a limitation of simple data types.** 

## Implementation
This is a golang implementation. The client side is made as an interceptor that decorates the streams' `grpc.ClientStream`, overriding the method `RecvMsg`. 

A client simply initiate it's client connection with an interceptor `grpcConst.StreamClientInterceptor`.

A convenience method `grpcConst.HeaderSetConstant` can be used to construct the header that can be sent using your server-side `stream.SendHeader` before sending messages. 

see [examples](/examples)

## Testing the overhead
This is tested vs. the gRPC example [`route_guide.proto`](examples/route_guide/proto/route_guide.proto).
In this example the unmarshalling and figuring the fields to set takes about 1 µs. Handling these (two) values on each message takes 60 ns (whether you set a value or not). When no header is sent, the overhead of the clientStream is 6 ns pr. message (tested on my local pc).

Locally using this specification does not make much sense (for a local connection using this package is about 14% slower), but this should help you if your network/ infrastructure is network I/O limited.

## TODO
- more tests
- break even? (how much data should the stream send to break even)
  - document the proto data overhead
  - test in real-world example: non-local servers
- chaining
- test proto2
- test non-nil structs in header
- break up merge code
