# gRPC constant

`grpcConst` is a slim reflection based tool to populate the client side data of gRPC calls without having to send that data in every call. This is useful to reduce the data transported between services that is the same for all objects in a stream. 

An example could be for spatiotemporal data where you may send several thousand data points that all share the same location, because a user queried a specific datatype (e.g temperature), at a specific location. And want to have a timeseries displayed. This data may share location, maybe a location name and you could have other fields that are added as convenience to the data,  but the data is constant across the request. 

## Specification
For a given `message` returned by an RPC. The server sends a defaulting proto marshal'ed base64 URLencoded header `x-grpc-const` this object of the same type of the `message` will be the default values for the client side to add to all messages. 

The client side does the most of the work, decoding, unmarshaling and padding the `message` with this data. 

## Overriding
Any `message` sent with a value in the same place as the default constant `message` will override the default.  You cannot override the default by setting the value the 0, null, empty string or empty list as they are considered null, you cannot set these values as default,  and that wouldn't make sense a they are default already.

## Implementation
This is a golang implementation. The client side is made as an interceptor that decorates the streams' `grpc.ClientStream`, overriding the method `RecvMsg`. 

A client simply initiate it's client connection with an interceptor `grpcConst.StreamClientInterceptor`.

A convenience method `grpcConst.HeaderSetConstant` can be used to construct the header that can be sent using your server-side `stream.SendHeader` before sending messages. 
## TODO
- arrays?
- test
- benchmark
- chaining
