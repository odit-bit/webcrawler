package crawler

/*
	WIP, changing API will still occure

	this package contain implementation logic of web crawler with data pipeline process.
	there is some important component:

	xpipe
		pipeline concurrent pattern that separate data operation by stage.

	rest
		JSON api protocol implementation (REST API)  for receive and sent data to be crawled

	rpc
		GRPC api protocol , WIP

	crawler.go
		a service that wire pipeline together from source , stage, sink

	resource.go
		Domain , data type represent url that will crawl

	html_*.go
		implement xpipe processorFunc, for each stage operation

*/
