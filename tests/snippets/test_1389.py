res = 0
for x in range(9):
    if x % 2 == 0:
        res += x
    else:
        res *= 2
print(res)
