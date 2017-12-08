Quickstart
----------
```
logger := spantool.LoggerDatadog{StatsdAddr: "datadog-agent.datadog:8125", Tags: []string{"service:micro-rating"}, PollingDuration: 5 * time.Second}
if err := spantool.LogSessionsCount(spannerClient, logger); err != nil {
	return nil, errors.Wrap(err, 0)
}
```
