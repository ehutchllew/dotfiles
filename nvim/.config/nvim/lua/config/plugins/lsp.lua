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
			opts.capabilities = capabilities
			local lspconfig = require("lspconfig")
			lspconfig.gopls.setup(opts)
			lspconfig.lua_ls.setup(opts)
			lspconfig.ts_ls.setup(opts)
			lspconfig.buf_ls.setup(opts)

			-- NOTE: Since I'm using conform I don't need this for now.
			-- NOTE: Leaving as documentation for now.
			-- vim.api.nvim_create_autocmd('LspAttach', {
			-- 	callback = function(args)
			-- 		local client = vim.lsp.get_client_by_id(args.data.client_id)
			-- 		if not client then return end
			--
			-- 		if client.supports_method("textDocument/formatting") then
			-- 			-- Format the current buffer on save
			-- 			vim.api.nvim_create_autocmd("BufWritePre", {
			-- 				buffer = args.buf,
			-- 				callback = function()
			-- 					vim.lsp.buf.format({ bufnr = args.buf, id = client.id })
			-- 				end,
			-- 			})
			-- 		end
			-- 	end,
			-- })
		end,
	}
}
