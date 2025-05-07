package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"

	"github.com/elhambadri2411/social/internal/store"
)

var usernames = []string{
	"Olivia", "Liam", "Emma", "Noah", "Ava", "Oliver", "Sophia", "Elijah", "Isabella",
	"Lucas", "Mia", "Mason", "Charlotte", "Logan", "Amelia", "Ethan", "Harper", "James",
	"Evelyn", "Aiden", "Abigail", "Jackson", "Emily", "Sebastian", "Ella", "Owen", "Elizabeth",
	"Samuel", "Camila", "Henry", "Luna", "Alexander", "Sofia", "Michael", "Avery", "Daniel",
	"Mila", "Jacob", "Aria", "Jack", "Scarlett", "Luke", "Penelope", "William", "Layla",
	"Joshua", "Chloe", "Matthew", "Victoria", "Joseph", "Madison", "Mateo",
}

var titles = []string{
	"Mastering Go Concurrency: A Deep Dive",
	"Understanding Go Interfaces and Type Embedding",
	"Building RESTful APIs with Go and Gin",
	"Exploring Go’s Garbage Collector Internals",
	"Effective Error Handling Patterns in Go",
	"Introduction to System Design: Fundamentals and Best Practices",
	"Scaling Databases: Sharding vs Replication",
	"A Comprehensive Guide to Networking Protocols",
	"Implementing Distributed Tracing with OpenTelemetry",
	"Demystifying Containerization with Docker",
	"Kubernetes Operators: Extending the API",
	"Performance Tuning in High-Throughput Systems",
	"Deep Learning Basics: From Neurons to Networks",
	"Reinforcement Learning from Scratch",
	"Applying LSM Trees in Modern Databases",
	"Bloom Filters: Probabilistic Data Structures Explained",
	"External Sorting Algorithms for Big Data",
	"Building a URL Shortener Service in Go",
	"Secure Authentication with OAuth2 and JWT",
	"Exploring the TCP/IP Stack: Step-by-Step",
	"Practical Guide to HTTP/2 and gRPC",
	"Mastering SQL Joins: Inner, Outer, and Beyond",
	"NoSQL Databases: When and How to Use Them",
	"Caching Strategies with Redis and Memcached",
	"Event-Driven Architectures with Kafka",
	"Real-Time Data Processing with Apache Flink",
	"Building CLI Tools in Go: Best Practices",
	"Testing Go Code: Unit, Integration, and E2E Tests",
	"Profiling Go Applications for Maximum Performance",
	"Advanced Concurrency Patterns: Workers, Pipelines, and Pools",
	"Implementing Circuit Breakers and Bulkheads",
	"Load Testing Your APIs with k6",
	"Securing Microservices with mTLS",
	"Understanding BGP and Internet Routing",
	"DNS Deep Dive: Resolution and Caching",
	"TLS Handshake: How HTTPS Works",
	"Container Networking with CNI Plugins",
	"Building Scalable Pub/Sub Systems",
	"Deep Dive into ACID and BASE Theories",
	"Two-Phase Commit vs Paxos vs Raft",
	"Observability: Metrics, Logs, and Traces",
	"Designing Idempotent APIs",
	"Building Real-Time Chat with WebSockets",
	"Go Memory Model: What Every Developer Should Know",
	"gRPC vs REST: Choosing the Right Tool",
	"Implementing Feature Flags in Production",
	"Blue-Green Deployments with Spinnaker",
	"GitOps: Automating Deployments with Flux",
	"Infrastructure as Code with Terraform",
	"Continuous Delivery Pipelines with Jenkins",
	"GraphQL vs REST: Pros and Cons",
	"Serverless Architectures: FaaS Explained",
}

var contents = []string{
	"Concurrency is the heart of Go's design, enabling efficient multi-tasking with goroutines and channels; in this deep dive, we explore advanced patterns and pitfalls.",
	"Go's interfaces and type embedding provide powerful abstractions for building flexible, composable code; we demonstrate how to design and implement them effectively.",
	"The Gin framework offers an elegant way to build high-performance RESTful APIs in Go; this guide walks you through setup, routing, middleware, and deployment.",
	"Go's garbage collector ensures memory safety and performance; this post peels back the curtain to reveal how its concurrent mark-and-sweep algorithm works under the hood.",
	"Error handling in Go is explicit and idiomatic; we cover patterns like sentinel errors, error wrapping, and custom error types to write robust applications.",
	"System design is a cornerstone of scalable architecture; this introductory article covers core concepts like load balancing, caching, and monitoring.",
	"When scaling databases, sharding and replication offer distinct trade-offs; we compare their architectures, use cases, and operational challenges.",
	"From Ethernet to HTTP/2, networking protocols dictate how data flows across the internet; here’s a structured overview of the most essential protocols.",
	"Distributed tracing helps you understand request flows across microservices; learn how to instrument your services with OpenTelemetry for end-to-end visibility.",
	"Docker revolutionized deployment with containerization; this post covers images, containers, networking, and best practices for production-ready setups.",
	"Operators bring custom automation to Kubernetes; discover how to build and deploy operators to manage complex applications at scale.",
	"High-throughput systems demand careful tuning; we explore profiling, bottleneck analysis, and optimization strategies to achieve peak performance.",
	"Deep learning has transformed AI; this primer introduces neural networks, activation functions, and training algorithms to get you started.",
	"Reinforcement learning enables agents to learn from interaction; we implement key algorithms like Q-learning and Policy Gradients step by step.",
	"Log-Structured Merge Trees power write-heavy databases; we examine their design, compaction strategies, and practical implementation considerations.",
	"Bloom filters offer space-efficient set membership tests with allowable false positives; learn their math, variants, and real-world uses.",
	"When data exceeds memory, external sorting is essential; this article covers merge sort-based techniques and I/O optimization tips.",
	"A URL shortener combines web APIs, databases, and hashing; follow this tutorial to build and deploy a scalable short link service in Go.",
	"OAuth2 and JWT form the backbone of modern auth; we dissect their flows, token management, and how to secure your APIs.",
	"Understanding TCP/IP is fundamental for network programming; we explore each layer, packet structure, and common protocols in detail.",
	"HTTP/2 and gRPC deliver high-performance RPC and streaming; learn how to configure protocols, handle multiplexing, and optimize for speed.",
	"SQL joins power data analysis; this comprehensive guide explains inner, left, right, full, and cross joins with practical examples.",
	"NoSQL databases offer flexible schemas and scalability; we compare document, key-value, and graph stores to help you choose the right solution.",
	"Caching reduces latency and database load; learn patterns like cache-aside, write-through, and eviction policies with Redis and Memcached.",
	"Kafka enables scalable event streaming; this post covers producers, consumers, topics, partitions, and how to build resilient event-driven systems.",
	"Flink offers powerful stream processing capabilities; discover its architecture, APIs, and windowing strategies for low-latency analytics.",
	"Go's simplicity and static binaries make it ideal for CLI tools; we cover Cobra, Viper, and tips for building user-friendly commands.",
	"Robust testing ensures code quality; learn how to write unit tests, integration tests, and end-to-end tests using Go’s testing package and tools.",
	"Profiling identifies performance hotspots; this guide shows how to use pprof, trace, and benchmarking to optimize Go programs.",
	"Beyond basic goroutines, advanced patterns like worker pools and pipelines enable complex concurrent workflows; see examples and use cases.",
	"Circuit breakers and bulkheads protect distributed systems from cascading failures; we implement patterns in Go and discuss tuning parameters.",
	"k6 is a modern load testing tool; learn how to write test scripts, simulate traffic, and analyze performance metrics to validate your APIs.",
	"Mutual TLS ensures both client and server authentication; this post explains certificate management and secure communication between services.",
	"BGP is the protocol that makes the internet work; we explore route advertisements, policies, and security challenges like hijacking.",
	"DNS translates domain names to IPs; learn about recursive resolvers, authoritative servers, and caching mechanisms in this detailed guide.",
	"HTTPS secures web traffic via TLS; this article breaks down the handshake, certificate validation, and cipher suites step by step.",
	"Container networking relies on CNI plugins; we compare popular options, explain network namespaces, and walk through plugin configuration.",
	"Pub/Sub architectures enable decoupled communication; this guide covers message brokers, topic design, and delivery guarantees for scale.",
	"Transaction models influence consistency and availability; explore ACID, BASE, and how modern databases apply these principles under the CAP theorem.",
	"Distributed consensus ensures data integrity; compare two-phase commit, Paxos, and Raft algorithms with code examples and trade-offs.",
	"Observability is key to reliable systems; learn how to collect, store, and analyze metrics, logs, and traces for full-stack visibility.",
	"Idempotency prevents unintended side effects; this post shows how to design and implement idempotent endpoints to handle retries safely.",
	"WebSockets enable bidirectional communication; learn to set up a real-time chat server in Go with authentication and message broadcasting.",
	"Understanding the Go memory model is crucial for safe concurrency; we cover happens-before, atomic operations, and memory barriers.",
	"gRPC and REST have distinct strengths; compare their performance, interoperability, and use cases to choose the right API paradigm.",
	"Feature flags allow controlled rollouts; learn how to integrate feature flag libraries and design safe rollout and rollback strategies.",
	"Blue-green deployments minimize downtime; this tutorial walks through Spinnaker pipelines and environment management for zero-downtime releases.",
	"GitOps brings Git-centric operations to the forefront; discover Flux, declarative configs, and best practices for version-controlled deployments.",
	"Terraform makes infrastructure repeatable and versioned; this guide covers modules, state management, and workspaces for team collaboration.",
	"Jenkins remains a staple for CI/CD; learn how to create pipelines, integrate tests, and manage credentials securely.",
	"GraphQL offers flexible querying while REST remains ubiquitous; compare their schemas, performance implications, and developer DX.",
	"Serverless enables event-driven functions; explore FaaS platforms, cold starts, and architecture patterns for cost-effective scalability.",
}

var tags = []string{
	"go",
	"concurrency",
	"system-design",
	"networking",
	"databases",
	"performance",
	"security",
	"microservices",
	"distributed-systems",
	"api",
	"testing",
	"cli",
	"observability",
	"containerization",
	"kubernetes",
	"cloud",
	"devops",
	"machine-learning",
	"streaming",
	"auth",
}

var comments = []string{
	"Great post!",
	"Very informative.",
	"Thanks for sharing!",
	"Helpful content.",
	"Nice overview.",
	"Well written.",
	"Good read.",
	"Valuable insights.",
	"Clear and concise.",
	"Learned something new.",
	"Appreciate the breakdown.",
	"This was useful.",
	"Fantastic article.",
	"Excellent write-up.",
	"Practical advice.",
	"Nice summary.",
	"Really helpful.",
	"Quality content.",
	"Very well explained.",
	"Good examples.",
	"Insightful post.",
	"Strong coverage.",
	"Concise and clear.",
	"Great introduction.",
	"Helpful tips.",
	"Useful guide.",
	"Keep up the good work.",
	"Looking forward to more.",
	"This helped a lot.",
	"Awesome post.",
}

func Seed(store store.Storage, db *sql.DB) error {
	ctx := context.Background()

	users := generateUsers(100)
	tx, _ := db.BeginTx(ctx, nil)

	for _, user := range users {
		if err := store.UsersRepository.Create(ctx, tx, user); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	tx.Commit()

	posts := generatePosts(200, users)
	for _, post := range posts {
		if err := store.PostsRepository.Create(ctx, post); err != nil {
			return err
		}
	}

	comments := generateComments(500, users, posts)
	for _, comment := range comments {
		if err := store.CommentsRepository.Create(ctx, comment); err != nil {
			return err
		}
	}

	log.Println("Seeding complete")
	return nil
}

func generateUsers(num int) []*store.User {
	users := make([]*store.User, num)

	for i := 0; i < num; i++ {
		var password store.Password
		password.Set("password")
		users[i] = &store.User{
			Username: usernames[i%len(usernames)] + fmt.Sprintf("%d", i),
			Email:    usernames[i%len(usernames)] + fmt.Sprintf("%d", i) + "@mail.com",
			Password: password,
		}
	}

	return users
}

func generatePosts(num int, users []*store.User) []*store.Post {
	posts := make([]*store.Post, num)

	for i := 0; i < num; i++ {
		user := users[rand.Intn(len(users))]
		index := rand.Intn(len(titles))

		posts[i] = &store.Post{
			UserId:  user.ID,
			Title:   titles[index],
			Content: contents[index],
			Tags: []string{
				tags[rand.Intn(len(tags))],
				tags[rand.Intn(len(tags))],
				tags[rand.Intn(len(tags))],
			},
		}
	}

	return posts
}

func generateComments(num int, users []*store.User, posts []*store.Post) []*store.Comment {
	generatedComments := make([]*store.Comment, num)
	for i := 0; i < num; i++ {
		generatedComments[i] = &store.Comment{
			PostId:  posts[rand.Intn(len(posts))].ID,
			UserId:  users[rand.Intn(len(users))].ID,
			Content: comments[rand.Intn(len(comments))],
		}
	}

	return generatedComments
}
