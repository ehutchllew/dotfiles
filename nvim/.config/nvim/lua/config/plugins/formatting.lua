return {
	"stevearc/conform.nvim",
	event = { "BufReadPre", "BufNewFile" },
	opts = {
		formatters_by_ft = {
			css = { "prettierd" },
			-- NOTE: not sure if I actually need the `go` one, seems to work fine without
			-- go = { "gofmt" },
			html = { "prettierd" },
			javascript = { "prettierd" },
			javascriptreact = { "prettierd" },
			json = { "prettierd" },
			yaml = { "prettierd" },
			typescript = { "prettierd" },
			typescriptreact = { "prettierd" },
		},
		format_on_save = {
			async = false,
			lsp_fallback = true,
			timeout_ms = 500
		}
	},
	config = function(_, opts)
		local conform = require("conform")

		conform.setup(opts)

		-- In normal mode the format will work on entire buffer
		-- In visual mode it will work on selected range
		vim.keymap.set({ "n", "v" }, "<space>lf", function()
			conform.format({
				async = false,
				lsp_fallback = true,
				timeout_ms = 500
			})
		end, { desc = "Format File or Range" })
	end
}
