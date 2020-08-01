# ssh-secret-disperser
This repository provides the solution to one of my interview tasks. Below is the description of the problem:


> Problem statement
> We generate a list of secrets everyday. There is an ssh server where users can ssh into it and read those values. Before connecting to the server they already know which secret they are going to use for that day by providing an index in the list of secrets (number between 1-10). 
They are able to get the secret only if:
both of them are connected at the same time 
they enter the same index of the secret. 
Each user has it’s own time based one time password (we can use google authenticator application).
The ssh dialog can look like this:
Please choose a number between 1 to 10:
Please enter OTP code:
Once both users are authenticated and connected at the same time, the ssh server closes the connections and prints the secret for that day.
The ssh connection should timeout after 2 minutes (in case the other user doesn't connect). When the first user logs in, the other user has a window of 2 minutes to do the same. The task is to implement the ssh server.
We would like to see this implementation in golang. You can use the default go ssh package and otp library. The users should be able to use a regular ssh client.


# Solution Details
- The ssh server listens to `0.0.0.0:2022`
- There are only two hardcoded users (`user1` and `user2`) in the system (see `user/storage.go`)
  - OTP secrets for demo users: `L2NJFMNEJRBI2SNZQ2HUJNGRDCZEGTGM` and `LQDE6HPJHG55LXAHQ4LNEN2J2G6UIFHC`
- I am using a self-generated ssh-rsa key pairs (in the root of the repo)
- Some modest tests can be found in `challenge/server_test.go`
- Run the server by building/running `main.go`, e.g., go run main.go
- Connect via any ssh client, e.g., `ssh -l user1 -p 2022 0.0.0.0`
  - ******NOTE******: For some reasons, on OSX the ssh command-line client fails when attempting to connect via remote notation, 
  i.e., `ssh user1@0.0.0.0:2022` fails but `ssh -l user1 -p 2022 0.0.0.0` works fine
  
  
  
