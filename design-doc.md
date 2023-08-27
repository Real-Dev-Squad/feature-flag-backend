## Title
Approaches for Rate limiting for Feature flag service [Design doc]
## Authors
Shubham Yadav, Vikhyat
## Problem 
We wanted to limit our AWS costs after a certain amount of requests, say the allowed request count is 500, we should not incur the costs. 
## Solution approaches
### Solution 1
#### Switching-off the lambda, once request quota is reached
- Things used in this solution are Dynamodb, one lambda for changing the config [To mark the reserved concurrency to zero]
- Steps in this solution
  - When a request comes to FF service, we first check if the value stored in DDB table is greater than 0
  - If yes we continue processing the request and return the requested API response
  - If the value in the database has become 0 then we make call to the new lambda    
