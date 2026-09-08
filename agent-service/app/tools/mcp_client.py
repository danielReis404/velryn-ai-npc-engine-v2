import os
import sys
from langchain_mcp_adapters.client import MultiServerMCPClient

_MCP_BASE_DIR = os.path.abspath(
    os.path.join(os.path.dirname(__file__), "..", "..", "..", "mcp")
)
if sys.platform.startswith("win"):
    _MCP_PYTHON_EXE = os.path.join(_MCP_BASE_DIR, ".venv", "Scripts", "python.exe")
else:
    _MCP_PYTHON_EXE = os.path.join(_MCP_BASE_DIR, ".venv", "bin", "python")
_MCP_SERVER_PATH = os.path.join(_MCP_BASE_DIR, "src", "velryn_mcp", "server.py")
_client = MultiServerMCPClient(
    {
        "velryn-actions": {
            "command": _MCP_PYTHON_EXE,
            "args": [_MCP_SERVER_PATH],
            "transport": "stdio",
        }
    }
)
_cached_tools = None


async def get_action_tools():
    global _cached_tools
    if _cached_tools is None:
        _cached_tools = await _client.get_tools()
    return _cached_tools
