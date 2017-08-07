/*
A time series database optimized for performance measurements.

Storing samples

Let's say you measure application latency several times per second.
Each sample is a JSON document:

	{
		"read_latency": 12.3
	}

To persist measurements, send the following HTTP request:

	curl -X POST http://localhost:8080/mydatabase -d '{"read_latency":12.3}'

where:

  `mydatabase` is time series database name. It's recommended to create a separate database for each benchmark.

Obviously, you will rather use your favourite programming language to send HTTP requests.

It's absolutely OK to create thousands of databases.

Aggregation and visualization

This API returns JSON document with aggregated characteristics (mean, percentiles, and etc.):

	$ curl -s http://127.0.0.1:8080/mydatabase/read_latency/summary | python -m json.tool
	{
		"avg": 5.82248,
		"count": 200000,
		"max": 100,
		"min": 0,
		"p50": 3,
		"p80": 9,
		"p90": 14,
		"p95": 21,
		"p99": 40,
		"p99.9": 76
	}

Please notice that Python is used for demonstration purpose only.

Finally, it is possible to generate heat map graphs in SVG format (use your browser to view):

	http://127.0.0.1:8080/mydatabase/read_latency/heatmap

Each rectangle is a cluster of values. The darker color corresponds to the denser population. The legend on the right side of the graph (the vertical bar) should help to understand the density.

Browsing data

To list all available database, use the following request:

	$ curl -s http://127.0.0.1:8080/ | python -m json.tool
	[
		"mydatabase"
	]

To list all metrics, use request similar to:

	$ curl -s http://127.0.0.1:8080/mydatabase | python -m json.tool
	[
		"read_latency",
		"write_latency"
	]

Querying samples

Only bulk queries are supported, but even they are not recommended.

To get the list of samples, use request similar to:

	$ curl -s http://127.0.0.1:8080/mydatabase/read_latency | python -m json.tool

Output is a JSON document with all timestamps and values:

	[
		[
			1437137708114018208,
			10
		],
		[
			1437137708114967597,
			15
		],
		[
			1437137708123781628,
			16
		]
	]

The first value in the nested list is the timestamp (the number of milliseconds elapsed since January 1, 1970 UTC).

The second value is the stored measurement (integer or float).
*/
package main
