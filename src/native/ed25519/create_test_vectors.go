package main

import (
    "crypto/sha512"
    "hash"
    "fmt"

    "github.com/dedis/kyber"
    "github.com/dedis/kyber/group/edwards25519"
    "github.com/dedis/kyber/sign/cosi"
    "github.com/dedis/kyber/util/random"
)

type Suite interface {
    kyber.Group
    kyber.Encoding
    kyber.XOFFactory
}

// Specify cipher suite using AES-128, SHA512, and the Edwards25519 curve.
type cosiSuite struct {
    Suite
}

func (m *cosiSuite) Hash() hash.Hash {
    return sha512.New()
}

var testSuite = &cosiSuite{edwards25519.NewBlakeSHA256Ed25519()}

// Since our implementation expands the private keys only when needed, we have to account for that.
func expand(private kyber.Scalar) (kyber.Scalar) {
    h := testSuite.Hash()
    private.MarshalTo(h)
    hash := h.Sum(nil)
    hash[0] &= 248
    hash[31] &= 63
    hash[31] |= 64
    m := testSuite.Scalar().Zero()
    m.UnmarshalBinary(hash[:32])
    return m
}

// The Bytes() method for Scalars reverses the array.
func reverse(bytes []byte) []byte {
    for i := 0; i < len(bytes)/2; i++ {
        j := len(bytes) - i - 1
        bytes[i], bytes[j] = bytes[j], bytes[i]
    }
    return bytes
}

// Example of using Schnorr
func main() {
    // Parameters begin
    message := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.")
    n := 5
    // Parameters end
    
    fmt.Printf("{\n")
    // Generate key pairs
    var privates []kyber.Scalar
    var publics []kyber.Point
    for i := 0; i < n; i++ {
        // kp := key.NewKeyPair(testSuite)
        private := testSuite.Scalar().Pick(random.Stream)
        public := testSuite.Point().Mul(expand(private), nil)
        privates = append(privates, private)
        publics = append(publics, public)
    }

    fmt.Printf("    privKeys: [")
    for i := 0; i < n; i++ {
        if i != 0 {
            fmt.Printf(", ")
        }
        fmt.Printf("BufferUtils.fromHex(\"%032x\")", reverse(privates[i].Bytes()))
    }
    fmt.Printf("],\n")

    fmt.Printf("    pubKeys: [")
    for i := 0; i < n; i++ {
        if i != 0 {
            fmt.Printf(", ")
        }
        fmt.Printf("BufferUtils.fromHex(\"%032s\")", publics[i].String())
    }
    fmt.Printf("],\n")

    // Init masks
    var masks []*cosi.Mask
    var byteMasks [][]byte
    for i := 0; i < n; i++ {
        m, err := cosi.NewMask(testSuite, publics, publics[i])
        if err != nil {
            panic(err.Error())
        }
        masks = append(masks, m)
        byteMasks = append(byteMasks, masks[i].Mask())
    }

    // Compute commitments
    var v []kyber.Scalar // random
    var V []kyber.Point  // commitment
    for i := 0; i < n; i++ {
        x, X := cosi.Commit(testSuite, nil)
        v = append(v, x)
        V = append(V, X)
    }

    fmt.Printf("    secrets: [")
    for i := 0; i < n; i++ {
        if i != 0 {
            fmt.Printf(", ")
        }
        fmt.Printf("BufferUtils.fromHex(\"%032x\")", reverse(v[i].Bytes()))
    }
    fmt.Printf("],\n")

    fmt.Printf("    commitments: [")
    for i := 0; i < n; i++ {
        if i != 0 {
            fmt.Printf(", ")
        }
        fmt.Printf("BufferUtils.fromHex(\"%032s\")", V[i].String())
    }
    fmt.Printf("],\n")

    // Aggregate commitments
    aggV, aggMask, err := cosi.AggregateCommitments(testSuite, V, byteMasks)
    if err != nil {
        panic(err.Error())
    }
    fmt.Printf("    aggCommitment: BufferUtils.fromHex(\"%032s\"),\n", aggV.String())

    // Set aggregate mask in nodes
    for i := 0; i < n; i++ {
        masks[i].SetMask(aggMask)
    }

    // Compute challenge
    var c []kyber.Scalar
    for i := 0; i < n; i++ {
        ci, err := cosi.Challenge(testSuite, aggV, masks[i].AggregatePublic, message)
        if err != nil {
            panic(err.Error())
        }
        c = append(c, ci)
    }
    fmt.Printf("    aggPubKey: BufferUtils.fromHex(\"%032s\"),\n", masks[0].AggregatePublic.String())

    // Compute responses
    var r []kyber.Scalar
    for i := 0; i < n; i++ {
        ri, _ := cosi.Response(testSuite, expand(privates[i]), v[i], c[i])
        r = append(r, ri)
    }

    fmt.Printf("    partialSignatures: [")
    for i := 0; i < n; i++ {
        if i != 0 {
            fmt.Printf(", ")
        }
        fmt.Printf("BufferUtils.fromHex(\"%032x\")", reverse(r[i].Bytes()))
    }
    fmt.Printf("],\n")

    // Aggregate responses
    aggr, err := cosi.AggregateResponses(testSuite, r)
    if err != nil {
        panic(err.Error())
    }
    fmt.Printf("    aggSignature: BufferUtils.fromHex(\"%032x\"),\n", reverse(aggr.Bytes()))

    for i := 0; i < n; i++ {
        // Sign
        sig, err := cosi.Sign(testSuite, aggV, aggr, masks[i])
        if err != nil {
            panic(err.Error())
        }
        // Verify (using default policy)
        if err := cosi.Verify(testSuite, publics, message, sig, nil); err != nil {
            panic(err.Error())
        }
    }

    fmt.Printf("    signature: BufferUtils.fromHex(\"%032s%032x\"),\n", aggV.String(), reverse(aggr.Bytes()))
    fmt.Printf("    message: BufferUtils.fromAscii(\"%032s\")\n", message)
    fmt.Printf("}\n")
}
