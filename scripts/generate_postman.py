import json
import re
from uuid import uuid4

ROUTES_FILE = 'routes/routes.go'

# mapping of variable name to group path
prefixes = {'app': ''}
items = []

method_regex = re.compile(r'(\w+)\.(Get|Post|Put|Delete)\("([^"]+)"')

with open(ROUTES_FILE) as f:
    for line in f:
        line = line.strip()
        # match group definitions
        match_group = re.match(r'(\w+)\s*:=\s*(\w+)\.Group\("([^"]+)"', line)
        if match_group:
            var, parent, path = match_group.groups()
            prefixes[var] = prefixes.get(parent, '') + path
            continue
        match_route = method_regex.search(line)
        if match_route:
            var, method, path = match_route.groups()
            full_path = prefixes.get(var, '') + path
            item = {
                "name": f"{method} {full_path}",
                "request": {
                    "method": method.upper(),
                    "header": [{"key": "Content-Type", "value": "application/json"}],
                    "url": {
                        "raw": f"{{{{base_url}}}}{full_path}",
                        "host": ["{{base_url}}"],
                        "path": [p for p in full_path.lstrip('/').split('/') if p]
                    }
                }
            }
            if method.upper() in ['POST', 'PUT']:
                item['request']['body'] = {"mode": "raw", "raw": "{}"}
            items.append(item)

collection = {
    "info": {
        "name": "Go Fiber API",
        "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
        "_postman_id": str(uuid4())
    },
    "item": items
}

with open('postman_collection.json', 'w') as f:
    json.dump(collection, f, indent=2)
