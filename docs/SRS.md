# Software requirements specification for project “Terraform provider for VirtualBox”

## Authors
- Barsukova Daria
- Petrov Vladimir
- Diza Mihail
- Mescheryakova Anastasia


## Introduction

>The main goal of our project is to write a provider for Terraform and then add new features to it. Providers are plugins that do all the work of interacting with resources. They should contain a description of the main properties of the resource, as well as the logic of CRUD operations (C - Create, R - Read, U - Update, D - Delete).  In our case we are going to manage the resources of virtual machines, which are quite a lot. For example, memory, number of processors, image of the system to be installed, etc. We want to implement a provider that can fully utilize the resources of virtual machines running on VIrtualBox (and later possibly VMWare Workstation). Terraform was chosen for a reason, it includes a clear declarative language, which is not difficult to learn. The main advantage of the declarative approach is that we do not need to analyze the state of the system every time, we just specify the state we want to get at the end, which simplifies the work. Of course, Terraform is not the only solution, like many other things it has analogs, but most of them do not have such a comfortable declarative approach in use.


## Glossary

- **Terraform** -  Open Source solution for managing IaC (Infrastructure as Code) from Hashicorp, released in 2014. Terraform uses a declarative style. It means that the user describes the final state of the infrastructure in the configuration file, and Terraform brings it to this state. Its main feature is the ability to expand the toolkit by installing additional modules. In Terraform terminology — **"providers"**.

- **Provider** - plugin for Terraform that can be used to manage the environment, virtual machines, hosts, clusters, inventory, network, data warehouses, content libraries and many others.

- **VirtualBox** - tool for virtualizing x86 and AMD64/Intel64 computing architecture for business and personal use.

- **Virtual machine** - compute resource that uses software instead of a physical computer to run programs and deploy apps.

- **DevOps** - set of practices, tools, and a cultural philosophy that automate and integrate the processes between software development and IT teams.

- **Go** - open source programming language that makes it simple to build secure, scalable systems.

- **Primary resources of virtual machines** - CPU, memory, network, and hard disk.

*  **Config files (.tf files)** - files that define the result we want, that is the state of the system that we want to get after the terraform work.


## Actors

> [!info] DevOps engineer
> **Goal**: Structured control of virtual machines.
**Precondition**: Terraform, go and virtualbox are installed and also have access to the internet. 
**Responsibilities**: has to describe the configuration of each virtual machine. (write config files)
>>
 > >>[!example] Main ideas of writing config files.
 > >>* **Write version.tf **- file that contains the versions and locations of the providers used. and also the terraform version.
 > >>* **Write main.tf **-  file that contains configurations of the wanted virtual machines, that is, specifying each resource like memory, CPUs, system image, boot method, etc.
 > >>*  **Write output.tf** - file that can be used to check some values that the user is interested in.

> [!info] Terraform
> **Goal**: Bring the system state to the state described in the configuration file.
> **Precondition**: Terraform, go and virtualbox are installed and also have access to the internet.
> **Responsibilities**: Compare the state of the system and the states described in the configuration file.

> [!info] VirtualBox
>**Goal**:  Working with virtual machines.
>**Precondition**: VirtualBox are installed.
>**Responsibilities**: Give the ability to manage virtual machines.

>[!info] QA
>**Goal**:  Good project.
>**Responsibilities**: Debugging and searching for errors in the code.

## Functional requirements

### Use case
> [!tip] Automated Management of VM
> **Actor**: DevOps engineer
> **Goal**: Avoid repetitive actions when using, deleting, creating and updating virtual machines, which helps to remove the human factor and possible errors.
>> [!todo] Main success scenario
>> - DevOps creates a configuration file.
>> - Runs Terraform.

>[!tip] Home use
>**Actor**: End User (often uses different VMs)
>**Goal**: Systematize work with updating and allocating virtual machine resources.
>>[!todo] Main success scenario
>>* User creates a configuration file or use the example.
>>* Runs Terraform.


### Stakeholders
- DevOps engineers
- VM and Terraform users
- some business??

### Environment
- OS - Windows 10 + / MacOS 12 + / Linux 6 +
- Terraform and GO on PC

## Non-functional requirements

### Performance
> The computer must be powerful enough to run virtual machines in a VirtualBox.
> Your computer must have enough memory to run on the os you want.

### Ease of support
>Users don't need to remember the state of the virtual machines yourself, the provider will bring the system to the desired state.
>The provider has good logging, which makes it easy to track errors and add new functionality.

### Extensibility
>A provider can increase its functionality by adding support for VMWare Workstation.

### Reliability
> Terraform regularly compares the state of the system with the configuration file, that avoids the human factor.











