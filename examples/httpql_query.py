# nPython Example: HTTPQL (Builder API) Query

def query_repo_info(token):
    # Using the builder-style HTTP API (HTTPQL)
    with scope("HTTP-ENV", token):
        with_client()
        set_url("https://api.github.com/repos/agenthands/npython")
        set_method("GET")
        resp = send_request()
        
        status = check_status(resp)
        if status == 200:
            print("Successfully queried repo info")
            # resp is a map, its 'body' field contains the raw data
            # But wait, our CheckStatus pops the map. 
            # So we need to store resp if we want to use it further.
            # In nPython, variables work as expected.
        else:
            print("Failed to query repo info, status: " + format_string("%s", status))

# Run the query
query_repo_info("token")
