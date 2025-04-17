# Run instruction for Golang Broadcast Chat

```bash
make deps
make build
make run
```
Run these 3 commands in one terminal -- server will open up

Then open up another terminal instance and do:

```bash
telnet localhost 8080
```
You can use this same command to run multiple users
Then follow the instr on the terminal

Please watch the video - its about 18 minutes and very fun to code along with.

----------------------------------------------------

# Run it on your local labs
## (My own stuff)
Start the server up just as mentioned in above
Make sure your local lab computers are on the same network
Then run the client folder's main file

```bash
cd client
go run main.go
```

The server periodically sends a broadcast to discover all clients trying to connect.
Once you connect, follow the terminal instr!
