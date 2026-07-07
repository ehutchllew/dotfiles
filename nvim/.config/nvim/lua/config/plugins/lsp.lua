return {
	{
		"neovim/nvim-lspconfig",
		dependencies = {
			"saghen/blink.cmp",
			{
				"folke/lazydev.nvim",
				ft = "lua", -- only load on lua files
				opts = {
					library = {
						-- See the configuration section for more details
						-- Load luvit types when the `vim.uv` word is found
						{ path = "${3rd}/luv/library", words = { "vim%.uv" } },
					},
				},
			},
		},
		opts = {},
		config = function(_, opts)
			local capabilities = require("blink.cmp").get_lsp_capabilities()
			vim.lsp.config("*", {
				capabilities = capabilities
			})

			local tsserver_filetypes = { "typescript", "javascript", "typescriptreact", "javascriptreact", "svelte" }
			local svelte_plugin = {
				name = "typescript-svelte-plugin",
				location = vim.fn.stdpath("data") ..
				"/mason/packages/svelte-language-server/node_modules/typescript-svelte-plugin",
				languages = { "svelte" },
				configNamespace = "typescript",
			}
			local vtsls_config = {
				settings = {
					vtsls = {
						tsserver = {
							globalPlugins = {
								svelte_plugin
							}
						}
					}
				},
				filetypes = tsserver_filetypes,
			}

			vim.lsp.config("vtsls", vtsls_config)

			vim.lsp.enable("gopls")
			vim.lsp.enable("lua_ls")
			vim.lsp.enable("vtsls")
			vim.lsp.enable("buf_ls")
			vim.lsp.enable("svelte")
			vim.lsp.enable("yamlls")
			vim.lsp.enable("zls")
		end,
	}
}
