import os
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer


class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        body = f"hello {os.environ.get('AUTHOR', '')}".encode("utf-8")
        self.send_response(200)
        self.send_header("Content-Type", "text/plain; charset=utf-8")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, format, *args):
        pass


port = int(os.environ.get("PORT", "8000"))
ThreadingHTTPServer(("0.0.0.0", port), Handler).serve_forever()
