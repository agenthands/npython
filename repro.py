def first_five_primes():
    primes = []
    num = 2
    while len(primes) < 5:
        is_prime = True
        # print("Checking:", num)
        for i in range(2, int(num**0.5) + 1):
            if num % i == 0:
                is_prime = False
                break
        if is_prime:
            primes.append(num)
            print("Found prime:", num)
        num += 1
    print(primes)

first_five_primes()
