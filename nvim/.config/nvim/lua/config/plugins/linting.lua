return {
	"mfussenegger/nvim-lint",
	event = {
		"BufReadPre",
		"BufNewFile"
	},
	opts = {
		linters_by_ft = {
			javascript = { "eslint_d" },
			javascriptreact = { "eslint_d" },
			typescript = { "eslint_d" },
			typescriptreact = { "eslint_d" },
		}
	},
	config = function(_, opts)
		local lint = require("lint")
		lint.linters_by_ft = opts.linters_by_ft

		local lint_augrp = vim.api.nvim_create_augroup("lint", { clear = true })

		vim.api.nvim_create_autocmd({ "BufEnter", "BufWritePost", "InsertLeave" }, {
			group = lint_augrp,
			callback = function()
				lint.try_lint()
			end
		})

		vim.keymap.set("n", "<space>ll", function()
			lint.try_lint()
		end, { desc = "Trigger Linting for Buffer" })
	end
}
