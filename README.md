# WebSpark
WebSpark is the web front server for IBM Spark cloud services. It enables Spark cloud instances to serve as web endpoints.

WebSpark is transparent to web browser and HTTP clients. Web clients receive standard HTTP responses, which are generated by Spark cloud instances. Otherwise, these Spark cloud instances do not normally accept inbound HTTP connections.

Spark cloud instances declare URLs they serve to WebSpark. These URLs are transparently mapped to web client requests. This eliminates the need for task queue support for web clients and Spark instances.

Special thanks to IBM and IBM Spark Hackathon for making this possible.

David Wang
