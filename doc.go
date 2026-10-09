// Package diesis gives Go programs typed access to the Diesis system
// contracts.
//
// It has the chain definitions and system contract addresses from
// canonical.json, a table pairing each fixed-address contract with its
// generated ABI, and small helpers that parse those ABIs and bind them with
// go-ethereum's accounts/abi/bind/v2 package. The generated ABIs, signatures,
// selectors, event topics, and parameter, event, error, and tuple structs live
// in the contracts subpackage.
package diesis
