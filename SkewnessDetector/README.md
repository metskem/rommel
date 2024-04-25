### SkewnessDetector

Listens on 2 firehoses for events and count them to detect skewness in the number of events between the 2 firehoses. 

Environment variables used:

* API_ADDR - a comma-separated list of 2 addresses of the Cloud Foundry API (https://api.sys.mydomain.com)
* CF_USERNAME - a comma-separated list of 2 usernames to use for the Cloud Foundry API
* CF_PASSWORD - a comma-separated list of 2 passwords to use for the Cloud Foundry API
* CHAT_ID - the Telegram chat id to send the alerts to
* BOT_TOKEN - the Telegram bot token to use to send the alerts
* THRESHOLD - the threshold for the skewness detection, absolute number of total requests (to prevent false alerts during low activity)
* THRESHOLD_PERC - the threshold for the skewness detection, percentage of skewness
* INTERVAL - the interval in seconds at which the skewness will be checked
