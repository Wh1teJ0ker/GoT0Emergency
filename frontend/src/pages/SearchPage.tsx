import { Search as SearchIcon } from 'lucide-react';
import { PageContainer } from "../components/layout/PageContainer";
import { PageHeader } from "../components/layout/PageHeader";
import { Input } from "../components/ui/Input";
import { Card, CardContent } from "../components/ui/Card";
import { Button } from "../components/ui/Button";

export function SearchPage() {
    return (
        <PageContainer>
            <PageHeader 
                title="Search" 
                description="Find projects, documents, and resources across your workspace."
            />

            <Card>
                <CardContent className="pt-6">
                    <div className="flex gap-4">
                        <div className="relative flex-1">
                            <SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground h-4 w-4" />
                            <Input 
                                type="text" 
                                placeholder="Type to search..." 
                                className="pl-9"
                            />
                        </div>
                        <Button>Search</Button>
                    </div>
                </CardContent>
            </Card>

            <div className="text-center py-12 text-muted-foreground">
                <p>Enter a search term to get started.</p>
            </div>
        </PageContainer>
    );
}
