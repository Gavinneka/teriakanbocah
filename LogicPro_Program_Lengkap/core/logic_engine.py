import re
import itertools

class LogicEngine:
    OPERATORS = {
        'NOT': ['!', '~', '¬'],
        'AND': ['&', '∧', '*', 'AND'],
        'OR': ['|', '∨', '+', 'OR'],
        'IF': ['->', '→', '=>'],
        'IFF': ['<->', '↔', '<=>']
    }

    PRECEDENCE = {
        'NOT': 4,
        'AND': 3,
        'OR': 2,
        'IF': 1,
        'IFF': 1
    }

    def __init__(self):
        pass

    def tokenize(self, expression):
        # Normalize expression
        expr = expression.upper()
        # Replace multi-char operators with single unique identifiers for processing
        expr = expr.replace('<->', ' IFF ').replace('<=>', ' IFF ').replace('↔', ' IFF ')
        expr = expr.replace('->', ' IF ').replace('=>', ' IF ').replace('→', ' IF ')
        
        # Add spaces around operators and parentheses
        for ops in self.OPERATORS.values():
            for op in ops:
                if op not in ['IF', 'IFF', 'AND', 'OR', 'NOT']:
                    expr = expr.replace(op, f' {op} ')
        
        expr = expr.replace('(', ' ( ').replace(')', ' ) ')
        
        # Split and clean
        tokens = expr.split()
        
        # Map tokens to internal types
        final_tokens = []
        for token in tokens:
            found = False
            for op_type, ops in self.OPERATORS.items():
                if token in ops or token == op_type:
                    final_tokens.append(('OP', op_type))
                    found = True
                    break
            if not found:
                if token == '(':
                    final_tokens.append(('LPAREN', '('))
                elif token == ')':
                    final_tokens.append(('RPAREN', ')'))
                elif re.match(r'^[A-Z]$', token):
                    final_tokens.append(('VAR', token))
                else:
                    raise ValueError(f"Invalid token: {token}")
        
        return final_tokens

    def to_rpn(self, tokens):
        """Convert tokens to Reverse Polish Notation using Shunting-yard algorithm."""
        output = []
        stack = []
        
        for type, val in tokens:
            if type == 'VAR':
                output.append((type, val))
            elif type == 'OP':
                while (stack and stack[-1][0] == 'OP' and 
                       self.PRECEDENCE[stack[-1][1]] >= self.PRECEDENCE[val]):
                    output.append(stack.pop())
                stack.append((type, val))
            elif type == 'LPAREN':
                stack.append((type, val))
            elif type == 'RPAREN':
                while stack and stack[-1][0] != 'LPAREN':
                    output.append(stack.pop())
                if not stack:
                    raise ValueError("Mismatched parentheses")
                stack.pop() # Remove LCAREN
        
        while stack:
            if stack[-1][0] == 'LPAREN':
                raise ValueError("Mismatched parentheses")
            output.append(stack.pop())
            
        return output

    def evaluate(self, rpn, var_values):
        stack = []
        for type, val in rpn:
            if type == 'VAR':
                stack.append(var_values[val])
            elif type == 'OP':
                if val == 'NOT':
                    a = stack.pop()
                    stack.append(not a)
                else:
                    b = stack.pop()
                    a = stack.pop()
                    if val == 'AND':
                        stack.append(a and b)
                    elif val == 'OR':
                        stack.append(a or b)
                    elif val == 'IF':
                        # A -> B is !A or B
                        stack.append((not a) or b)
                    elif val == 'IFF':
                        stack.append(a == b)
        return stack[0]

    def generate_truth_table(self, expression):
        try:
            tokens = self.tokenize(expression)
            rpn = self.to_rpn(tokens)
            variables = sorted(list(set(val for type, val in tokens if type == 'VAR')))
            
            if not variables:
                return None, "No variables found"

            header = variables + [expression]
            rows = []
            
            # Combinations of True/False
            for combo in itertools.product([True, False], repeat=len(variables)):
                var_map = dict(zip(variables, combo))
                result = self.evaluate(rpn, var_map)
                rows.append(list(combo) + [result])
                
            return {"header": header, "rows": rows, "variables": variables}, None
        except Exception as e:
            return None, str(e)

if __name__ == "__main__":
    # Quick test
    engine = LogicEngine()
    table, err = engine.generate_truth_table("P -> Q")
    if err:
        print(f"Error: {err}")
    else:
        print(table['header'])
        for row in table['rows']:
            print(row)
