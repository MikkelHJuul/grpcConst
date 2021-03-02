# gRPC constant

`grpcConst` is a slim reflection based tool to populate the client side data of gRPC messages without having to send that data in every message. This is useful to reduce the data transported between services where the stream messages share data. 

An example could be for spatiotemporal data where you may send several thousand data points that all share the same location, because a user queried a specific datatype (e.g temperature), at a specific location. And want to have a timeseries displayed. This data may share location, maybe a location name and you could have other fields that are added as convenience to the data, but those fields are constant across the request. 

## Specification
Client side sends a header of key `x-grpc-const` this signals the server that it implements this feature. 

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

For the full automatic experience on the server-side wrap your stream using `grpcConst.ServerStreamWrapper` to reduce the default data before sending your messages

see [examples](/examples)

## Testing the overhead
This is tested vs. the gRPC example [`route_guide.proto`](examples/route_guide/proto/route_guide.proto).
In this example the unmarshalling and figuring the fields to set takes about 1 µs. Handling these (two) values on each message takes 45 ns (whether you set a value or not). When no header is sent, the overhead of the clientStream is 7 ns pr. message (tested on my local pc).
```
BenchmarkDataAddingClientStream_RecvMsgSimple           # This merges two values into an object
BenchmarkDataAddingClientStream_RecvMsgSimple-8       	15383833            67.5 ns/op
BenchmarkDataAddingClientStream_RecvMsgNil              # This doesn't initiate anything
BenchmarkDataAddingClientStream_RecvMsgNil-8          	86265670            14.2 ns/op
BenchmarkDataAddingClientStream_RecvMsgNeverWrite       # This checks for the values but never writes them
BenchmarkDataAddingClientStream_RecvMsgNeverWrite-8   	16704519            67.8 ns/op
BenchmarkDataAddingClientStream_RecvMsgLarger           # This merges 12 default values
BenchmarkDataAddingClientStream_RecvMsgLarger-8       	5098196             231 ns/op
BenchmarkInitiation                                     # Initiation and merge one message
BenchmarkInitiation-8                                   239176              4349 ns/op
```
Locally using this specification does not make much sense, but this should help you if your network/ infrastructure is network I/O limited.
As a side-note (anecdotal evidence), when in a video-conference the local clients that uses this library (and one that does not) starts performing more evenly (to see this example check [`example/ogc_ish`](examples/ogc_ish)). <br>
To compare with [`imdario/mergo`](https://github.com/imdario/mergo) (which is not at all fair, as the focus is on safety and configurability, not performance) a benchmark on my computer, says (this is the same result as `BenchmarkInitiation`):
```
BenchmarkMergo
BenchmarkMergo-8   	  159506	      6740 ns/op
```
## TODO
- possibly write merger using protoreflect
- client interceptor with base Merger; implement protreflect.Merge as a merger
- benchmark reducer
