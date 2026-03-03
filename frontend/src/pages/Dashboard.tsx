import { PageContainer } from "../components/layout/PageContainer";
import { PageHeader } from "../components/layout/PageHeader";
import { Button } from "../components/ui/Button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../components/ui/Card";
import { Badge } from "../components/ui/Badge";
import { Plus } from "lucide-react";

export function Dashboard() {
    return (
        <PageContainer>
            <PageHeader 
                title="Dashboard" 
                description="Welcome to your workspace. Here is an overview of your activity."
                action={<Button><Plus className="mr-2 h-4 w-4" /> New Project</Button>}
            />
            
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {[1, 2, 3].map((i) => (
                    <Card key={i} className="hover:shadow-md transition-shadow">
                        <CardHeader>
                            <div className="flex justify-between items-start">
                                <CardTitle className="text-lg">Project Alpha {i}</CardTitle>
                                <Badge variant={i === 1 ? "default" : "secondary"}>
                                    {i === 1 ? "Active" : "Draft"}
                                </Badge>
                            </div>
                            <CardDescription>Last updated 2 hours ago</CardDescription>
                        </CardHeader>
                        <CardContent>
                            <p className="text-sm text-muted-foreground">
                                This is a sample card content to demonstrate the layout structure and component usage.
                            </p>
                            <div className="mt-4 flex gap-2">
                                <Button variant="outline" size="sm">View</Button>
                                <Button variant="ghost" size="sm">Edit</Button>
                            </div>
                        </CardContent>
                    </Card>
                ))}
            </div>
        </PageContainer>
    );
}
