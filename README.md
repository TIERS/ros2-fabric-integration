# ROS 2 + Hyperledger Fabric Blockchain

## Introduction

This work presents an integration of ROS 2 with the Hyperledger Fabric blockchain. With a framework that leverages Fabric smart contracts and ROS 2 through a Go application, we delve into the potential of using blockchain for controlling robots, and gathering and processing their data. 

You can refer to our [paper preprint](https://arxiv.org/abs/2203.03426) for more details. Please use this citation if the code in this repo is useful for your work:
```
@article{salimi2022dlt, 
      title="Towards Managing Industrial Robot Fleets with Hyperledger Fabric Blockchain and {ROS} 2", 
      author="Salma Salimi and Jorge {Pe\~na Queralta} and Tomi Westerlund", 
      journal="arXiv preprint", 
      publisher="arXiv", 
      year="2022"
}
```


## What is included in this repo?

Examples for Integrating ROS 2 and Hyperledger Fabric

[conceptual figure](fig/ros2fabric.png)

## What is the purpose?

Integration of ROS2 and Hyperledger Fabric. This package provides samples of smart contract and application which are all written in GO.

## Installation

Clone this repo 
```
git clone git@github.com:TIERS/ros2-fabric-integration.git
```

Before bringing up your network, replace this assetTransfer.go with your current application. Also replace this smartcontract.go with your current smartcontract in the chaincode-go file.

## Contact

Visit us at [https://tiers.utu.fi](https://tiers.utu.fi)

