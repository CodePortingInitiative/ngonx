package nginx

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// LineType represents the type of configuration line
type LineType string

const (
	LineTypeComment   LineType = "comment"   // Contains only comments
	LineTypeInclude   LineType = "include"   // Include directive
	LineTypeDirective LineType = "directive" // Regular directive
	LineTypeBlock     LineType = "block"     // Starts a new block
)

// Line represents a single line in the nginx configuration
type Line struct {
	Name     string   // Name of the directive
	Params   []string // Parameters for the directive
	Comments []string // Comments associated with this line
	Type     LineType // Type of the line
}

// Block represents a configuration block in nginx
type Block struct {
	Name      string   // Name of the block (e.g., "server", "http")
	Params    []string // Parameters for the block (e.g., "example.com" in "server example.com")
	Lines     []*Line  // Lines directly in this block
	Blocks    []*Block // Child blocks
	Comments  []string // Comments associated with this block definition
	ParentRef *Block   // Reference to parent block, nil for root
}

// Config represents the entire nginx configuration
type Config struct {
	RootBlock *Block // Root block of the configuration
	FilePath  string // Path to the configuration file
}

// ParseConfig parses the nginx configuration file
func ParseConfig(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	rootBlock := &Block{
		Name:   "root",
		Params: []string{},
		Lines:  []*Line{},
		Blocks: []*Block{},
	}

	config := &Config{
		RootBlock: rootBlock,
		FilePath:  filePath,
	}

	scanner := bufio.NewScanner(file)
	currentBlock := rootBlock
	blockStack := []*Block{rootBlock}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parseLine(line, currentBlock, &blockStack)

		// Update currentBlock to be the last block in the stack
		if len(blockStack) > 0 {
			currentBlock = blockStack[len(blockStack)-1]
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return config, nil
}

// parseLine processes a single line of nginx configuration
func parseLine(line string, currentBlock *Block, blockStack *[]*Block) {
	// Handle comments
	commentStart := strings.Index(line, "#")
	var comments []string

	if commentStart >= 0 {
		commentText := strings.TrimSpace(line[commentStart+1:])
		comments = []string{commentText}
		line = strings.TrimSpace(line[:commentStart])
	}

	// Skip if it's an empty line after removing comments
	if line == "" {
		if len(comments) > 0 {
			// This is a comment-only line
			currentBlock.Lines = append(currentBlock.Lines, &Line{
				Type:     LineTypeComment,
				Comments: comments,
			})
		}
		return
	}

	// Handle block end
	if line == "}" {
		// Pop the current block from stack
		if len(*blockStack) > 1 {
			*blockStack = (*blockStack)[:len(*blockStack)-1]
		}
		return
	}

	// Split by semicolons to handle multiple directives in one line
	directives := splitDirectives(line)

	for _, directive := range directives {
		directive = strings.TrimSpace(directive)
		if directive == "" {
			continue
		}

		// Parse the directive
		parts := strings.Fields(directive)
		if len(parts) == 0 {
			continue
		}

		name := parts[0]
		params := []string{}

		if len(parts) > 1 {
			params = parts[1:]
		}

		// Check if it's a block start
		if directive[len(directive)-1] == '{' {
			// This is a block directive
			blockName := parts[0]
			blockParams := []string{}

			// Remove the opening brace from the last param if it's there
			if len(parts) > 1 {
				lastParam := parts[len(parts)-1]
				if lastParam == "{" {
					blockParams = parts[1 : len(parts)-1]
				} else if strings.HasSuffix(lastParam, "{") {
					blockParams = parts[1 : len(parts)-1]
					blockParams = append(blockParams, strings.TrimSuffix(lastParam, "{"))
				} else {
					blockParams = parts[1:]
				}
			}

			// Create new block
			newBlock := &Block{
				Name:      blockName,
				Params:    blockParams,
				Lines:     []*Line{},
				Blocks:    []*Block{},
				Comments:  comments,
				ParentRef: currentBlock,
			}

			// Add to parent's blocks
			currentBlock.Blocks = append(currentBlock.Blocks, newBlock)

			// Push to stack
			*blockStack = append(*blockStack, newBlock)

			// Add block line to parent
			currentBlock.Lines = append(currentBlock.Lines, &Line{
				Name:     blockName,
				Params:   blockParams,
				Comments: comments,
				Type:     LineTypeBlock,
			})

			// Clear comments as they've been used
			comments = nil
		} else {
			// Regular directive or include
			lineType := LineTypeDirective
			if name == "include" {
				lineType = LineTypeInclude
			}

			currentBlock.Lines = append(currentBlock.Lines, &Line{
				Name:     name,
				Params:   params,
				Comments: comments,
				Type:     lineType,
			})

			// Clear comments as they've been used
			comments = nil
		}
	}
}

// splitDirectives splits a line by semicolons but respects quotes
func splitDirectives(line string) []string {
	var results []string
	var current strings.Builder
	inQuote := false
	quoteMark := rune(0)

	for _, char := range line {
		switch char {
		case '\'', '"':
			if inQuote && char == quoteMark {
				inQuote = false
				quoteMark = rune(0)
			} else if !inQuote {
				inQuote = true
				quoteMark = char
			}
			current.WriteRune(char)
		case ';':
			if inQuote {
				current.WriteRune(char)
			} else {
				results = append(results, strings.TrimSpace(current.String()))
				current.Reset()
			}
		case '{':
			current.WriteRune(char)
			if !inQuote {
				results = append(results, strings.TrimSpace(current.String()))
				current.Reset()
				return results
			}
		default:
			current.WriteRune(char)
		}
	}

	// Add any remaining content
	if current.Len() > 0 {
		results = append(results, strings.TrimSpace(current.String()))
	}

	return results
}

// PrintConfig prints the parsed configuration for debugging
func PrintConfig(config *Config) {
	printBlock(config.RootBlock, 0)
}

// printBlock recursively prints a block with proper indentation
func printBlock(block *Block, indent int) {
	indentStr := strings.Repeat("  ", indent)

	// Print block info
	if block.Name != "root" {
		blockStr := indentStr + block.Name
		if len(block.Params) > 0 {
			blockStr += " " + strings.Join(block.Params, " ")
		}
		blockStr += " {"
		if len(block.Comments) > 0 {
			blockStr += " # " + strings.Join(block.Comments, " ")
		}
		println(blockStr)
	}

	// Print lines
	for _, line := range block.Lines {
		if line.Type == LineTypeBlock {
			continue // Already handled when printing blocks
		}

		lineStr := indentStr + "  " + line.Name
		if len(line.Params) > 0 {
			lineStr += " " + strings.Join(line.Params, " ")
		}
		if line.Type != LineTypeComment {
			lineStr += ";"
		}
		if len(line.Comments) > 0 {
			lineStr += " # " + strings.Join(line.Comments, " ")
		}
		println(lineStr)
	}

	// Print child blocks
	for _, childBlock := range block.Blocks {
		printBlock(childBlock, indent+1)
	}

	// Print closing brace for non-root blocks
	if block.Name != "root" {
		println(indentStr + "}")
	}
}

// PrintTree prints the configuration as a hierarchical tree with detailed information
func (config *Config) PrintTree() {
	fmt.Printf("Configuration File: %s\n", config.FilePath)
	fmt.Println("└── Root")
	printTreeBlock(config.RootBlock, "    ")
}

// printTreeBlock prints a block as part of the tree representation
func printTreeBlock(block *Block, prefix string) {
	// Print lines
	for i, line := range block.Lines {
		isLast := i == len(block.Lines)-1 && len(block.Blocks) == 0

		// Choose the appropriate branch character
		branch := "├── "
		if isLast {
			branch = "└── "
		}

		// Format line representation
		lineInfo := formatLineInfo(line)

		fmt.Printf("%s%s%s\n", prefix, branch, lineInfo)

		// Print comments if any and not included in the line info
		if line.Type != LineTypeComment && len(line.Comments) > 0 {
			commentPrefix := prefix
			if isLast {
				commentPrefix += "    "
			} else {
				commentPrefix += "│   "
			}

			for j, comment := range line.Comments {
				commentBranch := "├── "
				if j == len(line.Comments)-1 {
					commentBranch = "└── "
				}
				fmt.Printf("%s%sComment: %s\n", commentPrefix, commentBranch, comment)
			}
		}
	}

	// Print blocks
	for i, childBlock := range block.Blocks {
		isLast := i == len(block.Blocks)-1

		// Choose the appropriate branch character
		branch := "├── "
		if isLast {
			branch = "└── "
		}

		// Format block representation
		blockInfo := childBlock.Name
		if len(childBlock.Params) > 0 {
			blockInfo += " " + strings.Join(childBlock.Params, " ")
		}
		blockInfo += " {}"
		if len(childBlock.Comments) > 0 {
			blockInfo += fmt.Sprintf(" (Comments: %d)", len(childBlock.Comments))
		}

		fmt.Printf("%s%sBlock: %s\n", prefix, branch, blockInfo)

		// Print block comments if any
		nextPrefix := prefix
		if isLast {
			nextPrefix += "    "
		} else {
			nextPrefix += "│   "
		}

		if len(childBlock.Comments) > 0 {
			for j, comment := range childBlock.Comments {
				commentBranch := "├── "
				if j == len(childBlock.Comments)-1 && len(childBlock.Lines) == 0 && len(childBlock.Blocks) == 0 {
					commentBranch = "└── "
				}
				fmt.Printf("%s%sComment: %s\n", nextPrefix, commentBranch, comment)
			}
		}

		// Recursively print child block content
		printTreeBlock(childBlock, nextPrefix)
	}
}

// formatLineInfo creates a string representation of a line with type and parameters
func formatLineInfo(line *Line) string {
	var result string

	switch line.Type {
	case LineTypeComment:
		result = fmt.Sprintf("Comment: %s", strings.Join(line.Comments, " "))
	case LineTypeInclude:
		result = fmt.Sprintf("Include: %s %s", line.Name, strings.Join(line.Params, " "))
	case LineTypeDirective:
		result = fmt.Sprintf("Directive: %s %s", line.Name, strings.Join(line.Params, " "))
	case LineTypeBlock:
		result = fmt.Sprintf("BlockStart: %s %s", line.Name, strings.Join(line.Params, " "))
	}

	return result
}
