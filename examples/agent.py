# nPython Example: Secure Data Pipeline

def run_agent(http_token, fs_token):
    # Fetch data and save it to a file
    with scope("HTTP-ENV", http_token):
        html = fetch("http://google.com")

    # String containment check
    if "google" in html:
        with scope("FS-ENV", fs_token):
            write_file(html, "google_home.html")
        print("Successfully saved Google home page.")
    else:
        print("Failed to find expected content.")
