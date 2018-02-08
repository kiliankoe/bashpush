#! /usr/bin/env python

import os
from sys import exit
import re
import json
import requests
import feedparser
from pushover_complete import PushoverAPI
import enforce


@enforce.runtime_validation
def env(var: str) -> str:
    """Read value from environment and exit if it doesn't exist."""
    val = os.getenv(var)
    if val is None:
        exit(f'missing env var {var}')
    return val


@enforce.runtime_validation
def fetchLatest() -> dict:
    """Fetch the newest quote from the RSS feed."""
    bash_url = env('BASH_FEED_URL')
    feed = requests.get(bash_url)
    # feedparser can't deal with the bash url directly 🙄
    latest = feedparser.parse(feed.text).entries[0]
    return {
        'url': latest['link'],
        'quote': latest['summary'],
        'id': int(re.findall(r'\?(\d+)', latest['link'])[0])
    }


@enforce.runtime_validation
def isNew(new_id: int) -> bool:
    """Check if the given quote id is newer than the last known one.
    If it is, it's persisted to `lastquote.txt`."""
    if not os.path.exists('lastquote.txt'):
        open('lastquote.txt', 'a').close()

    with open('lastquote.txt', 'r+') as f:
        last_id = f.read()
        last_id = 0 if last_id is '' else int(last_id)

        if new_id > last_id:
            f.seek(0)
            f.write(str(new_id))
            return True
        else:
            return False


@enforce.runtime_validation
def sendPushoverNotifications(latest: dict):
    users = env('PUSHOVER_USER_TOKENS').split(',')
    p = PushoverAPI(env('PUSHOVER_API_TOKEN'))
    for user in users:
        p.send_message(user, latest['quote'], url=latest['url'])


@enforce.runtime_validation
def sendSlackNotifcation(latest: dict):
    slack_url = env('SLACK_HOOK_URL')
    quote_str = f'Neues Bash Zitat 👉 <{latest["url"]}|{latest["id"]}> 🤐'
    quote_json = json.dumps({'text': quote_str})
    r = requests.post(slack_url, data=quote_json)
    if r.status_code is not 200:
        exit(f'Post to Slack failed for {latest["id"]} /o\\\n{r.text}')


if __name__ == '__main__':
    latest = fetchLatest()
    if isNew(latest['id']):
        sendSlackNotifcation(latest)
        sendPushoverNotifications(latest)
