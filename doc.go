/*
Fast, flexible and reliable storage for performance measurements.

Storing samples

Let's say you collect CPU stats every 5 seconds, each sample is represented by a JSON object or document:

    {
        "cpu_sys": 12.3,
        "cpu_user": 50.4,
        "cpu_idle": 37.3
    }

You can persist your measurements by sending the following HTTP request:

    $ curl -XPOST http://localhost:8080/mybenchmark/app1 -d @sample.json

Where:

"mybenchmark" is a common snapshot entity. You should change it before *any* test or benchmark iteration.

"app1" is a source name. In this case we are using application name, it can be an IP address (e.g., "172.23.100.96") or name of database (e.g., "mydatabase@127.0.0.1").

"sample.json" is the JSON document which we described above.

User-specified timestamps:

    $ curl -XPOST http://localhost:8080/mybenchmark/app1 -d @sample.json

Querying samples

It cannot be simpler:

    $ curl http://localhost:8080/mybenchmark/app1/cpu_sys

Output is a JSON document as well:

    {"1403736306507708119": 12.3}

where `1403736306507708119` is sample timestamp (the number of nanoseconds elapsed since January 1, 1970 UTC).

Listing snapshots sources and metrics

In order to list all snapshots:

    $ curl http://localhost:8080/

In order to list all sources for given snapshot:

    $ curl http://localhost:8080/mybenchmark

Getting a list of distinct metrics:

    $ curl http://localhost:8080/mybenchmark/app1

Summary and visualization

This API returns JSON document with aggregated metrics:

    $ curl http://localhost:8080/mybenchmark/app1/cpu_sys/summary

output:

    {
        "avg": 34.2,
        "max": 87.1,
        "min": 0.1,
        "p50": 37.3,
        "p80": 52.4,
        "p90": 72.5,
        "p95": 81.7,
        "p99": 85.2
    }

Built-in visualization using amazing D3 charts:

    $ http://localhost:8080/mybenchmark/app1/cpu_sys/linechart
*/
package main
