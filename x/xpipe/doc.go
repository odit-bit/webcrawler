package xpipe

/*
WIP

INTERFACE
	1. Fetcher
		interface that implemented by thing that can fetch data T
	2. Streamer
		stream data to consumer


TYPE
	1. Producer[T any]
		producer embed the Fetcher interface to fetch data from outside (network)

	2. Consumer[T any]
		consumer embed Streamer to stream every data recieved from data-bus(last stage) and error-bus (merge)

	3. Pipe
		implement fifo apporach of data pipeline pattern where data processed sequentialy from producer , stages and consumer

	4.

FUNCTION
	Stage()
		generic function that cant receive data T from upstream to proccesed and return the result <-chan T and the error

	Merge()
		generic function for merge multiple chan into one chanel (Fan-in). the very first purpose is to merge or create the error-bus, so Consumer can consume the error.
		But it not just for it, it can act as bridge for the next stage that received from many result channel of previous stages.

*/
