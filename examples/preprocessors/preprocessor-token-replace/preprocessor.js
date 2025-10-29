#!/usr/bin/env node

/**
 * GeoPub Token Replace Preprocessor
 * 
 * This is an example preprocessor that demonstrates the mdBook-compatible
 * preprocessor protocol. It reads a preprocessor context from stdin,
 * replaces tokens in chapter content, and writes the modified context to stdout.
 * 
 * Tokens are in the format {{TOKEN_NAME}}.
 * The replacement values come from the preprocessor config in book.toml.
 * 
 * Usage in book.toml:
 * [preprocessor.token-replace]
 * command = "node preprocessors/token-replace/preprocessor.js"
 * AUTHOR_NAME = "Jane Doe"
 * VERSION = "1.0.0"
 */

const fs = require('fs');

// Read JSON from stdin
function readStdin() {
  return new Promise((resolve, reject) => {
    let data = '';
    process.stdin.setEncoding('utf8');
    process.stdin.on('readable', () => {
      let chunk;
      while ((chunk = process.stdin.read()) !== null) {
        data += chunk;
      }
    });
    process.stdin.on('end', () => {
      try {
        resolve(JSON.parse(data));
      } catch (err) {
        reject(new Error(`Failed to parse stdin: ${err.message}`));
      }
    });
    process.stdin.on('error', reject);
  });
}

// Replace all token occurrences in a string
function replaceTokens(content, config) {
  let result = content;
  
  // Build regex from config keys
  for (const [key, value] of Object.entries(config)) {
    // Skip known non-token keys
    if (['command', 'renderers', 'before', 'after'].includes(key)) {
      continue;
    }
    
    const token = `{{${key}}}`;
    // Use global regex replace
    const regex = new RegExp(token.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'g');
    result = result.replace(regex, String(value));
  }
  
  return result;
}

// Recursively process chapters and replace tokens
function processChapter(chapter, config) {
  if (!chapter) return;
  
  // Replace tokens in content
  chapter.content = replaceTokens(chapter.content, config);
  
  // Process sub-items
  if (chapter.sub_items && Array.isArray(chapter.sub_items)) {
    for (const subItem of chapter.sub_items) {
      if (subItem.chapter) {
        processChapter(subItem.chapter, config);
      }
    }
  }
}

// Main preprocessor function
async function main() {
  try {
    // Read context from stdin
    const context = await readStdin();
    
    if (!context.book || !context.book.sections) {
      throw new Error('Invalid preprocessor context: missing book.sections');
    }
    
    // Build config dict from the preprocessor config
    // Look for this preprocessor's config (can be named anything, just get first one)
    const config = {};
    if (context.config && context.config.preprocessor) {
      // Get the first (or any) preprocessor config that isn't a known setting
      for (const [name, cfg] of Object.entries(context.config.preprocessor)) {
        // If it's an object with our settings, use it
        if (typeof cfg === 'object' && cfg !== null) {
          // This should be our preprocessor config
          Object.assign(config, cfg);
          break;
        }
      }
    }
    
    // Process all sections
    for (const section of context.book.sections) {
      if (section.chapter) {
        processChapter(section.chapter, config);
      }
    }
    
    // Write result to stdout
    process.stdout.write(JSON.stringify(context));
  } catch (err) {
    console.error(`Error: ${err.message}`);
    process.exit(1);
  }
}

main();
