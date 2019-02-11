# gotermchat
A Golang Chat Server with a Client Terminal. Chat with your fellow developers via terminal or command prompt!


All the speed, lightness and power of Golang and MongoDB are at your beck and call with this utility!

To use out of the box, setup MongoDB and create a database called `UserService`. In this database, 
create a collection called `Users`. Ensure that MongoDB is running on port 27017.

Navigate to the gms.go file located in gotermchat/cmd/startsrv and do: go build.

Then do: ./startsrv -p 8080 . (Where 8080 is the port you would love to start your server on)

Note the ip address of your system and then

On the client systems which you would like to chat from, also go to gotermchat/cmd/termclient. Do go build again.

Now do  <code>./termclient -username=Angel.Seraphim -h=127.0.0.1 -p=8080 -phone=0906678888 reg</code>. (Use your phone number!)

The `reg` is for a first time user. It ensures that the server sees you as a new user and so, registers you.
The 127.0.0.1 is the ip address of the server.
The 8080 is the port on which the server runs.

Subsequently, you can chat without login. Enjoy!
