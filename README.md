# dusk-go
A local branch of the reference implementation of Dusk. Please check out the whitepaper at: https://github.com/dusk-network/whitepaper/releases/download/v0.3/dusk-whitepaper.pdf.


## Creating a New Key Pair

```
    entropy,_ = crypto.RandEntropy(32)

    privateKey, _ := wallet.NewPrivateKey(entropy) // 7ea55e73267...

    publicKey := privateKey.Public() // d194f142b2e5...
```

## Generating the Public Address and Private WIF

```
    PublicAddress, _ := publicKey.PublicAddress() // DUSKpub1JBG1FrnwDwtaZnXP3z6NazsXzS3j9B5vBPhszfa3xDpLeCFQkj2M

    WIF, _:= privateKey.WIF() // DUSKpriv1WCxh37LxDgeLk45Khnyo9cdBG6NC3Ax12ZpGFgD8Mgd7KfapGDg
```

## Sign and Verify

```
    // Sign

    message := []byte("hello world")
    signature, _ := privateKey.Sign(message)

    // Verify

    valid := publicKey.Verify(message, signature) // true
```