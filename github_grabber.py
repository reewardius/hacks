import re
import urllib
import requests
import json

from urllib.parse import urlparse

GITHUB_API_KEY = r"ghp_[a-zA-Z0-9]{36}"

TOKEN_DATA = "YOUR_TOKEN_HERE"

def get_raw_file(file_url):
    
    raw_host = "raw.githubusercontent.com"
    
    github_original_host = urlparse(file_url).hostname

    result = file_url.replace(github_original_host, raw_host)

    return result.replace('blob/', '')


def find_match(raw_url):
    
    resource = urllib.request.urlopen(raw_url)
    content = resource.read().decode(resource.headers.get_content_charset())
    keys = re.findall(GITHUB_API_KEY, content)

    return keys


def github_api_search_code(qwery, page):

    raw_urls = []

    headers = {"Authorization": "token " + TOKEN_DATA}

    url = 'https://api.github.com/search/code?s=indexed&type=Code&o=desc&q=' + qwery + '&page=' + str(page)

    r = requests.get(url, headers=headers)

    response_json = json.loads(r.content)

    for item in response_json['items']:
        raw_urls.append(get_raw_file(item['html_url']))

    return raw_urls


if __name__ == '__main__':

    print("--- COLLECTING DATA ---")
    
    for url in github_api_search_code('ghp_', n):
        if find_match(url):
            print(find_match(url), ' found in ' + url)

    print("--- collecting data finished ---")
