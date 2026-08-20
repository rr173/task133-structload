package review
import("testing";"task133-structload/internal/model")
func TestBug28_ReviewRejectsEmptyComponentScope(t *testing.T){r:=Build(Input{ProjectID:"p",Combinations:[]model.LoadCombination{{ID:"combo"}}});if r.Ready{t.Fatal("review with no components was marked ready")};found:=false;for _,f:=range r.Findings{if f.Code=="no_components"{found=true}};if !found{t.Fatalf("findings=%+v",r.Findings)}}
