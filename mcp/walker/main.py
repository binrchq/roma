from mcp.server.fastmcp import FastMCP

# Initialize FastMCP server
mcp = FastMCP("walker")

@mcp.tool()
async def watch_ssh_window(session: str) -> str:
    """Watch ssh window content.

    Args:
        session: length 6 rand session id (e.g. exidk4,i8wssd)
    """
    text=f"""
       ______
      /\     \\
     />.\_____\\
   __\  /  ___/__        _ROMA__
  /\  \/__/\     \  ____/
 /O \____/*?\_____\\
 \  /    \  /     /                 [A seamless solution for remote access, ensuring both efficiency and security.]
  \/_____/\/_____/
commands:use In 1s whoami awk clear exit grep help history
agent.roma ~ help
use [OPTIONS]TYPE
Switch to specified TYPE of resource,TYPE is linux,router,windows,docker,database,switch,etc.
Usage:
-h,--help Display this help message
1n [-t TYPE]RESOURCE or RESOURCE
Login the specified TYPE of resource,TYPE is linux,router,windows,docker,database,switch;RESOURCE for ls Query,etc.
Usage:
-t,--type=TYPE Resource type
-h,--help
Display this help message
1s [OPTIONS]TYPE
List the specified TYPE of resource,TYPE is linux,router,windows,docker,database,switch,etc.
Usage:
-1,--list Display detailed information
-a,--all Display all resource
-h,--help Display this help message
whoami Get user(me)information
awk [OPTIONS]PATTERN ACTION
Process the input text according to the specified PATTERN and ACTION.
Usage:
-F,--field-separator-FIELD-SEPARATOR Specify the field separator
-h,--help
Display this help message
clearClear the screen
exit Exit the program
grep Search for PATTERN in input
help -Gets more help messages for commands
history Display command history
agent.roma ~
  """
    return "\n---\n".join(text)

def main():
    mcp.run(transport='stdio')


if __name__ == "__main__":
    main()
