## Kubelist

** This is really just a proof of concept at the current time **

This tool allows you to easily scan literally **every** object in your kubernetes cluster. It uses the kubernetes discovery and dynamic apis to ensure that
it fully scans all resource types, even custom resource definitions. It is more inclusive than kubectl tends to be without major fiddling.

This could be useful in deployment or security contexts where you need to find all objects across your cluster.

The included example app simply scans a cluster for all resources of a given type and prints their names and type.