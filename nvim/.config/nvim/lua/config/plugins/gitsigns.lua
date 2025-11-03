return {
	"lewis6991/gitsigns.nvim",
	opts = {
		current_line_blame = true
	},
	config = function(_, opts)
		local gs = require("gitsigns")
		gs.setup(opts)
	end
}
