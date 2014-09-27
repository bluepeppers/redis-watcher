Redis-Watcher
-------------

Watch given keys and metrics on a redis server and regularly reports
the results to a statsd server.

Running
-------

    redis-watcher -redis=127.0.0.1:6379 -statsd=127.0.0.1:8125

By default this will do nothing; you need to add some keys/metrics to
a...

Config file
-----------

### Basic command reporting

The original use case of redis-watcher was to regularly report the
sizes of queues (I use # for comment in the examples; remove them to
make a valid .json config file)

    {
        # By default, eval each metric every 3 seconds
        "report-interval": 3000
        "metrics": [
            # A simple metric might just be a name, and a command to run
            {
                "name": "crawler.requests.queue.length",
                "command": "zcard all:requests"
            },
            # However you can specify an interval yourself if you like
            {
                "name": "crawler.crawled.queue.length",
                "command": "llen scraped:urls",
                "report-interval": 5000
            }
        }
    }

Dump this in `/etc/redis-watcher.json` and when redis-watcher is run,
the `zcard all:requests` will be reported every 3 seconds.

By default, when a metric is run, the result (which is expected to be
a scalar) will be stored using a gauge in statsd. I would be happy to
remove both these limitations; I just haven't found a need to yet.

#### Internal reporting

You can configure redis-watcher to report the results of the `INFO`
redis command. This looks like

    {
        "internal": [
            # The key in this entry refers to the key in the
            # datastructure returned by INFO
            {
                "name": "redis.connected_clients",
                "key": "connected_clients"
            }
            
        ]
    }
