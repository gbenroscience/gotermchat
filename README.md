# gotermchat
A Golang Chat Server with a Client Terminal. Chat with your fellow developers via terminal or command prompt!


All the speed, lightness and power of Golang and MongoDB are at your beck and call with this utility!

To use out of the box, setup MongoDB and create a database called `UserService`. In this database, 
create a collection called `Users`. Ensure that MongoDB is running on port 27017.

Navigate to the gms.go file located in gotermchat/cmd/startsrv and do: go build.

Then do: <br><code>./startsrv -p 8080</code> </code>. (Where 8080 is the port you would love to start your server on)

Note the ip address of your system and then

On the client systems which you would like to chat from, also go to gotermchat/cmd/termclient. Do ``go build again``.

Now do <br> ``./termclient -u=Angel.Seraphim -h=127.0.0.1 -p=8080 -ph=0906678888 reg``.<br> (Use your phone number!)

`reg` is for a first time user. It ensures that the server sees you as a new user and so, registers you.
`127.0.0.1` is the ip address of the server.
`8080` is the port on which the server runs.

Subsequently, you can chat without login. Enjoy!

Modes of operation:

The server by default broadcasts to every one connected to it.
However, private messages can be sent even from the terminal by doing:

```<pr:080xxxxxx>``` where ```080xxxxxx``` is the phone number of the person you want to message.

The next milestone is group-chat.
