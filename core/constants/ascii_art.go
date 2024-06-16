package constants

// ASCIIArt is the ASCII art for the logo

const ROMA_ASCII01 = `
       ______
      /\     \
     />.\_____\
   __\  /  ___/__        _ROMA__
  /\  \/__/\     \  ____/
 /O \____/*?\_____\
 \  /    \  /     /                 [A seamless solution for remote access, ensuring both efficiency and security.]
  \/_____/\/_____/
`

const DOCKER_ASCII01 = `
            ##        .
        ## ## ##       ==
     ## ## ## ## ##    ===
/"""""""""""""""""\___/ ===
~~~ {~~ ~~~~ ~~~ ~~~~ ~~ ~ /  ===- ~~~
    \______ o          __/
      \    \        __/
        \____\______/

`
const DOCKER_ASCII02 = `
      ##         .
 ## ## ##        ==
## ## ## ## ##    ===
/""""""""""""""""\___/ ===
|     Docker     |   /  ===
\________________/_______|

`

const DOCKER_ASCII03 = `
         _.-^^---....,,--
     _--                  --_
    <                        >)
    |                         |
     \._                   _./
        '''--. . , ; .--'''
              | |   |
           .-=||  | |=-.
           '-=#$%&%$#=-'
              | ;  :|
     _____.,-#%&$@%#&#~,._____
     
`

const LINUX_ASCII01 = `
   .--.
  |o_o |
  |:_/ |
 //   \ \
(|     | )
/'\_   _/'\
\___)=(___/

`

const LINUX_ASCII02 = `
  .--.
 |o_o |
 |:_/ |
//   \ \
\     |/
 \_/\_/

`

const WINDOWS_ASCII01 = `
       _.-;;-._
'-..-'|   ||   |
'-..-'|_.-;;-._|
'-..-'|   ||   |
'-..-'|_.-''-._|

`
const WINDOWS_ASCII02 = `
${c1}                       .oodMMMM
                   .oodMMMMMMMMMMMMM
       ..oodMMM  MMMMMMMMMMMMMMMMMMM
 oodMMMMMMMMMMM  MMMMMMMMMMMMMMMMMMM
 MMMMMMMMMMMMMM  MMMMMMMMMMMMMMMMMMM
 MMMMMMMMMMMMMM  MMMMMMMMMMMMMMMMMMM
 MMMMMMMMMMMMMM  MMMMMMMMMMMMMMMMMMM
 MMMMMMMMMMMMMM  MMMMMMMMMMMMMMMMMMM
 MMMMMMMMMMMMMM  MMMMMMMMMMMMMMMMMMM
 
 MMMMMMMMMMMMMM  MMMMMMMMMMMMMMMMMMM
 MMMMMMMMMMMMMM  MMMMMMMMMMMMMMMMMMM
 MMMMMMMMMMMMMM  MMMMMMMMMMMMMMMMMMM
 MMMMMMMMMMMMMM  MMMMMMMMMMMMMMMMMMM
 MMMMMMMMMMMMMM  MMMMMMMMMMMMMMMMMMM
  ^^^^^^MMMMMMM  MMMMMMMMMMMMMMMMMMM
         ^^^^    ^^MMMMMMMMMMMMMMMMM
                          ^^^^^^MMMM
                      `
const DATABASE_ASCII01 = "" +
	"-. .-.   .-. .-.   .-. .-.   .\n" +
	"||\\|||\\ /|||\\|||\\ /|||\\|||\\ /|\n" +
	"|/ \\|||\\|||/ \\|||\\|||/ \\|||\\||\n" +
	"~   `-~ `-`   `-~ `-`   `-~ `-\n"

const SWITCH_ASCII01 = `
 _______________
|               |
|   Port 1      |
|   Port 2      |
|   Port 3      |
|   ...         |
|   Port 48     |
|_______________|

`
const SWITCH_ASCII02 = `
switch -\

`
const ROUTER_ASCII01 = `
 |__|__|
|  ___  |
|_______|

`

const ROUTER_ASCII02 = `
|_| roooooooooooouter

`

var AsciiPrompts = map[string][]string{
	ResourceTypeDocker:   {DOCKER_ASCII01, DOCKER_ASCII02, DOCKER_ASCII03},
	ResourceTypeLinux:    {LINUX_ASCII01, LINUX_ASCII02},
	ResourceTypeDatabase: {DATABASE_ASCII01},
	ResourceTypeSwitch:   {SWITCH_ASCII01, SWITCH_ASCII02},
	ResourceTypeRouter:   {ROUTER_ASCII01, ROUTER_ASCII02},
	ResourceTypeWindows:  {WINDOWS_ASCII01, WINDOWS_ASCII02},
	"~":                  {ROMA_ASCII01},
}
