# Pxier -- Simple but powerful free proxy API server
*Pxier* is an API server based on multiple source of proxy providers, and it's built with [Echo](https://github.com/labstack/echo), using a `MySQL` database to store all the data.It also provider a `report` api for user to report a dead proxy, therefore *Pxier* could improve the quality in real-time<br>
*Pxier*'s source code is so simply that you can easily fork or pull request. It's designed to keep everything simple. <br>
-- *It just a proxy "scraper", why bother doing those fascinating abstraction stuff?*

### Before start using *Pxier*
- *Pxier* only collect proxies, what you plan to do with *Pxier* is your freedom, not *Pxier*'s business, nor responsible to *Pxier* 
- *Pxier* is designed to provider an all-in-one solution, for those who being scraped and feel offended, please open an issue, and I'll delete your source right away
- *Pixer* is an individual project, since it's a scraper, it needs feedback to fix bugs, and more validated source for more "All-In-One"
- *Pixer* is an individual project, means that I may have no much time for fixing bug, develop more feature, so if you want to get more cool stuff, welcome to contribute or fork then do it yourself
- *Pixer* is an open source and free project, you're welcome to use, modify, publish your version. But since I don't know that deep about open source license and don't want to spend time for it, I think just put my words here is good enough: *If you planned to publish your own fork version, please name the origin author. If you planned to use in a commercial project, you MUST contact me first and MUST NOT use it before I agree.* 

## APIs
- `/status`: check *Pxier* status
- `/require`: require proxies
  - Param_`num`: how many proxies you want to require
  - Param_`provider`: which provider[Check here](./core/types.go) you want to use, split by `,`
- `/report`: mark a proxy as error, once a proxy's error times exceed the max error value, it will be deleted from the database
  - Param_`id`: Proxy's ID

### Install & Use

#### From Release
1. Download the suitable executable and config file and put them together
2. modify the config file
3. run the executable

#### From Docker
1. `docker pull ghcr.io/jobberrt/pxier:master`
2. check the [config.example.yaml](./config.example.yaml) and see if anything you need to change. If any, use -env to set(env key could be set to UPPERCASE, since I only test that).
3. run the docker container with your customized config(also expose the port)

#### Build yourself
1. `git clone the repo`
2. cd into the saved directory
3. `go mod tidy && go mod vendor && go build -o pixer`
4. `./pixer`

### Available Providers
Please check [File](./core/types.go)
