# Melody: Machine Edge Learning and Dynamic Scheduling 
Melody is an online scheduling framework for 
edge computing ML model inference on Kubernetes. 

## Overview
Melody uses Deep Q-learning as the base scheduling algorithm, use prometheus to monitor Edge computing resources and input metrics to DQN, and obtain the optimal scheduling strategy.
![Stack](../Downloads/iShot2022-06-19 19.06.53.png)
## Features
key benefits include:
- Support auto-scheduling ML inference pods between edge nodes.
- Continuously balance the computing resources (cpu, memory) on edge nodes.
- Consider the local model store for online serving jobs.
- Equipped with DQN reinforcement learning algorithm to obtain the balanced scheduling strategy.

# Custome Resource Definition(CRD)
- Inference define the ML inference jobs, it observes the scheduling decesion CRD, and dynamic adjust the resource limit and request. 
- Scheduling Decesion define the scheduling result from DQN learning algorithm server, it define the optimal CPU/memory resource for each inference.
## Get Started
