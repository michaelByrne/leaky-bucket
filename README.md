1) 

Build and run the app: 

docker build -t viaduct .

docker run -p 3000:3000 viaduct

2) 

The application uses gorilla mux to implement the limit endpoint. Gorilla provides routing capabilities, which we don't really need here, but it also allows us to easily implement middleware. Authorization functionality is provided as a middleware that essentially acts as a 'wrapper' around the endpoint (an endpoint that returns an endpoint, essentially), checking that an(y) API key is present on the incoming request. The decision to implement authorization as a middleware allows us to seperate the authorization logic from from the transport layer logic of the endpoint itself. 

A key design decision was in the selection of the algorithm itself: the leaky bucket algorithm, a variation of the token bucket algorithm. It was chosen for its relative implementation simplicity compared to the token bucket. In it, each user as identified by their API key is given a 'bucket' containing 10 'drips' or requests. Every time a user hits the endpoint, one drip leaves the bucket. If the bucket is empty, all requests have been exhausted per the given time period and the request returns a 429 error for rate limit exceeded. Each second, the bucket is refilled with an additional 10 total drips. This algorithm also happens to play very well with Go channels.  

A decision was made to not seperate the rate limiting logic into its own middleware. This was mostly a time consideration. Ideally, rate limiting would exist in its own layer, leaving the main driver file to simply declare the endpoints and set up the server. 

3) 
a) In a production system, for something this simple, I would likely deploy this endpoint in a Docker container to an EC2 instance, or multiple instances behind some kind of load balancer. 
b) Scaling this system would be a matter of running multiple instances of our application behind a load balancer, while Go's http library (which operates underneath Gorilla Mux) automatically uses goroutines to handle concurrency. That is, every time a user hits a handler function it happens in a new thread, leaving the main one unblocked. Further production-grade scaling could be accomplished using Kubernetes. 
c) For production-grade monitoring, we could use a tool like Prometheus, which allows us to embed listeners into our code that capture various metrics, e.g. counters, histograms, etc. Alternatively, we could also just log metrics if our server is to remain as simple as it currently is. Some interesting metrics we may want to capture are request rates per API key, how often rate limits are exceeded (per individual users and generally), when during the day rate limits are most likely to be exceeded, whether rate limit excesses are clustered together in time for multiple users (suggesting IP address spoofing is occurrent), what the relative proportions of new and recurring users are. 
