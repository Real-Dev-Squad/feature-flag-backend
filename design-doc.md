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
  - If the value in the database has become 0 then we make call to the new lambda to make the `reserved concurrency` to 0
  - This will make the API gateway not able to connect to the lambda `"Internal server error"`
- Costs incurred in this solution
  - Taking an example for 10000 requests made and the limit for the lambdas is 500 request
    - So for 500 requests the lambda costs is $0.
    - For rest 9500 requests the lambda costs is $0.04 rounding off to $0.5
  -  Now considering the dynamodb costs
  -  Dynamodb costs for 10000 requests is around $0.2, rounding this off to $0.5
  - Data transfer costs is around $0.09 **per GB** rounding off to $1
    - Considering the max data size = 10kb, so which means, for 100000 requests we will incur 0.09$ approx rounding it off to $1
  - API gateway costs for first 1 million request = $3.5
  
