# obliv_server.py
# Ismail Ahmed: An external HTTP server to allow the main Go code to use the oblivious HIRB data structure

# Needed imports
from http.server import BaseHTTPRequestHandler, HTTPServer
import json
from obliv.hirb import Hirb
from obliv.voram import create_voram

# Initalizing parameters (adjust as needed)
NUM_BLOBS = 128 # Default maximum number of blobs
BLOB_SIZE = 64  # Default bytes per blob, as it MUST MATCH THE BLOCK_SIZE VARIABLE IN PathORAM/include/blocks.h
NODE_SIZE = BLOB_SIZE * 12 # Default node size, MUST ALWAYS BE AT LEAST TWELVE TIMES BLOB_SIZE 
oram = create_voram(blobs_limit=NUM_BLOBS, blob_size=BLOB_SIZE, nodesize=NODE_SIZE)

# Creating the HIRB using the vORAM as a generic ORAM storage
store = Hirb(da_oram=oram, valsize=BLOB_SIZE)
class RequestHandler(BaseHTTPRequestHandler):
    # All our operations will be POST operations
    def do_POST(self):
        content_len = int(self.headers.get('Content-Length', 0))
        body = self.rfile.read(content_len)
        try:
            # Common values
            req = json.loads(body)
            op = req["op"]
            key = req["key"]
            val = req.get("val")

            # GET/READ JSON operation
            if op == "get":
                result = store.get(key, None)
            # SET/WRITE JSON operation
            elif op == "set":
                store[key] = val
                result = "ok"
            # DELETE/REMOVE operation
            elif op == "del":
                del store[key]
                result = "ok"
            # Invalid operations will have to be handled and output a HTTP 400 code
            else:
                self.send_response(400)
                self.end_headers()
                self.wfile.write(b"Invalid operation")
                return

            # Common process
            self.send_response(200)
            self.end_headers()
            self.wfile.write(json.dumps({"result": result}).encode())
        # Errors will have to be handled and output a HTTP 500 code
        except Exception as e:
            self.send_response(500)
            self.end_headers()
            self.wfile.write(str(e).encode())

# Running the HTTP server at Port 8236
def run(server_class=HTTPServer, handler_class=RequestHandler, port=8236):
    server = server_class(('', port), handler_class)
    print(f"[PythonAPI] Serving HIRB on port {port}")
    server.serve_forever()

# Running the HTTP server at Port 8236
if __name__ == "__main__":
    run()


# # Needed inputs
# from flask import Flask, request, jsonify
# from obliv.hirb import Hirb

# # Flask class definitions
# app = Flask(__name__)
# obliv_map = Hirb()

# # GET/READ JSON operation
# @app.route('/get', methods=['POST'])
# def get():
#     key = request.json['key']
#     value = obliv_map.get(key)
#     return jsonify({'value': value})

# # SET/WRITE JSON operation
# @app.route('/set', methods=['POST'])
# def set():
#     key = request.json['key']
#     value = request.json['value']
#     obliv_map[key] = value
#     return jsonify({'status': 'ok'})

# # DELETE/REMOVE operation
# @app.route('/delete', methods=['POST'])
# def delete():
#     key = request.json['key']
#     del obliv_map[key]
#     return jsonify({'status': 'deleted'})

# # Running at Port 8236
# if __name__ == '__main__':
#     app.run(port=8236)
