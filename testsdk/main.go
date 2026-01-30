package main

import (
	"context"
	"fmt"
	"log"

	"testsdk/api"
	"testsdk/api/fields"
	"testsdk/api/mutations"
	"testsdk/api/queries"
)

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}

func main() {
	client := api.NewClient("http://localhost:8080/graphql",
		api.WithAuthToken("your-token-here"),
	)

	ctx := context.Background()

	// Create query and mutation roots
	qr := queries.NewQueryRoot(client)
	mr := mutations.NewMutationRoot(client)

	// Example 1: Simple query
	pingResult, err := qr.Ping().Execute(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Ping result: %v\n", pingResult)

	// Example 2: Query chatbots with field selection
	chatbotsResult, err := qr.Chatbots().
		First(intPtr(10)).
		Select(func(conn *fields.ChatbotConnectionFields) {
			conn.TotalCount()
			conn.Edges(func(e *fields.ChatbotEdgeFields) {
				e.Cursor()
				e.Node(func(c *fields.ChatbotFields) {
					c.ID().
						Name().
						CreatedAt().
						UpdatedAt()
				})
			})
			conn.PageInfo(func(p *fields.PageInfoFields) {
				p.HasNextPage().
					HasPreviousPage().
					StartCursor().
					EndCursor()
			})
		}).
		Execute(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Chatbots: %+v\n", chatbotsResult)

	// Example 3: Query credentials with nested selection
	credentialsResult, err := qr.Credentials().
		First(intPtr(5)).
		Select(func(conn *fields.CredentialConnectionFields) {
			conn.TotalCount()
			conn.Edges(func(e *fields.CredentialEdgeFields) {
				e.Cursor()
				e.Node(func(c *fields.CredentialFields) {
					c.ID().
						Name().
						CreatedAt()
				})
			})
		}).
		Execute(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Credentials: %+v\n", credentialsResult)

	// Example 4: Query users
	usersResult, err := qr.Users().
		First(intPtr(10)).
		Select(func(conn *fields.UserConnectionFields) {
			conn.TotalCount()
			conn.Edges(func(e *fields.UserEdgeFields) {
				e.Cursor()
				e.Node(func(u *fields.UserFields) {
					u.ID().
						Email().
						FirstName().
						LastName().
						CreatedAt()
				})
			})
		}).
		Execute(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Users: %+v\n", usersResult)

	// Example 5: Query folders
	foldersResult, err := qr.Folders().
		First(intPtr(20)).
		Select(func(conn *fields.FolderConnectionFields) {
			conn.TotalCount()
			conn.Edges(func(e *fields.FolderEdgeFields) {
				e.Cursor()
				e.Node(func(f *fields.FolderFields) {
					f.ID().
						Name().
						Position().
						CreatedAt()
				})
			})
		}).
		Execute(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Folders: %+v\n", foldersResult)

	// Example 6: Chatbots with nested Users and AiModel
	chatbotsNestedResult, err := qr.Chatbots().
		First(intPtr(5)).
		Select(func(conn *fields.ChatbotConnectionFields) {
			conn.TotalCount()
			conn.Edges(func(e *fields.ChatbotEdgeFields) {
				e.Node(func(c *fields.ChatbotFields) {
					c.ID().
						Name().
						CreatedAt()
					// Nested: select AI model fields
					c.AiModel(func(ai *fields.AiModelFields) {
						ai.ID().
							ModelID().
							ProviderName().
							Status()
					})
					// Nested: select users
					c.Users(func(u *fields.UserFields) {
						u.ID().
							Email().
							FirstName().
							LastName()
					})
				})
			})
		}).
		Execute(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Chatbots with nested: %+v\n", chatbotsNestedResult)

	// Example 7: Folders with nested Chatbot, Owner, and Channels
	foldersNestedResult, err := qr.Folders().
		First(intPtr(10)).
		Select(func(conn *fields.FolderConnectionFields) {
			conn.TotalCount()
			conn.Edges(func(e *fields.FolderEdgeFields) {
				e.Node(func(f *fields.FolderFields) {
					f.ID().
						Name().
						Position()
					// Nested: folder owner
					f.Owner(func(u *fields.UserFields) {
						u.ID().
							Email().
							FirstName()
					})
					// Nested: parent chatbot
					f.Chatbot(func(c *fields.ChatbotFields) {
						c.ID().
							Name()
					})
					// Nested: channels in this folder
					f.Channels(func(ch *fields.ChannelFields) {
						ch.ID().
							Name().
							CreatedAt()
					})
				})
			})
		}).
		Execute(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Folders with nested: %+v\n", foldersNestedResult)

	// Example 8: Channels with deeply nested fields (3 levels)
	channelsResult, err := qr.Channels().
		First(intPtr(5)).
		Select(func(conn *fields.ChannelConnectionFields) {
			conn.TotalCount()
			conn.Edges(func(e *fields.ChannelEdgeFields) {
				e.Node(func(ch *fields.ChannelFields) {
					ch.ID().
						Name().
						CreatedAt()
					// Level 2: Channel owner
					ch.Owner(func(u *fields.UserFields) {
						u.ID().
							Email()
					})
					// Level 2: Channel's chatbot
					ch.Chatbot(func(c *fields.ChatbotFields) {
						c.ID().
							Name()
						// Level 3: Chatbot's AI model
						c.AiModel(func(ai *fields.AiModelFields) {
							ai.ID().
								ModelID().
								ProviderName()
						})
					})
					// Level 2: Channel's folder
					ch.Folder(func(f *fields.FolderFields) {
						f.ID().
							Name().
							Position()
					})
				})
			})
		}).
		Execute(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Channels with deep nesting: %+v\n", channelsResult)

	// Example 9: Mutation - HTML to Markdown
	markdownResult, err := mr.HTMLToMarkdown().
		Input(api.HtmlToMarkdownInput{
			HTML: "<h1>Hello World</h1><p>This is a test</p>",
		}).
		Execute(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Markdown: %+v\n", markdownResult)
}
